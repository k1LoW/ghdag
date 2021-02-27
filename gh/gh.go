package gh

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/k1LoW/ghdag/target"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const limit = 100

type Client struct {
	v3    *github.Client
	v4    *githubv4.Client
	owner string
	repo  string
}

// NewClient return Client
func NewClient() (*Client, error) {
	ctx := context.Background()

	// GITHUB_TOKEN
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_TOKEN")
	}

	// REST API Client
	v3c := github.NewClient(httpClient(token))
	if v4ep := os.Getenv("GITHUB_API_URL"); v4ep != "" {
		baseEndpoint, err := url.Parse(v4ep)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(baseEndpoint.Path, "/") {
			baseEndpoint.Path += "/"
		}
		v3c.BaseURL = baseEndpoint
	}

	// GraphQL API Client
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	v4hc := oauth2.NewClient(ctx, src)
	v4ep := os.Getenv("GITHUB_GRAPHQL_URL")
	if v4ep == "" {
		v4ep = "https://api.github.com/graphql"
	}
	v4c := githubv4.NewEnterpriseClient(v4ep, v4hc)

	ownerrepo := os.Getenv("GITHUB_REPOSITORY")
	if ownerrepo == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_REPOSITORY")
	}
	splitted := strings.Split(ownerrepo, "/")

	return &Client{
		v3:    v3c,
		v4:    v4c,
		owner: splitted[0],
		repo:  splitted[1],
	}, nil
}

type issueNode struct {
	Author struct {
		Login githubv4.String
	}
	Number       githubv4.Int
	Title        githubv4.String
	Body         githubv4.String
	URL          githubv4.String
	CreatedAt    githubv4.DateTime
	UpdatedAt    githubv4.DateTime
	LastEditedAt githubv4.DateTime
	Labels       struct {
		Nodes []struct {
			Name githubv4.String
		}
	} `graphql:"labels(first: 100)"`
	Assignees struct {
		Nodes []struct {
			Login githubv4.String
		}
	} `graphql:"assignees(first: 100)"`
	Comments struct {
		Nodes []struct {
			Author struct {
				Login githubv4.String
			}
			Body      githubv4.String
			CreatedAt githubv4.DateTime
		}
		PageInfo struct {
			HasNextPage bool
		}
	} `graphql:"comments(first: $limit, orderBy: {direction: DESC, field: UPDATED_AT})"`
}

type pullRequestNode struct {
	Author struct {
		Login githubv4.String
	}
	Number         githubv4.Int
	Title          githubv4.String
	Body           githubv4.String
	URL            githubv4.String
	IsDraft        githubv4.Boolean
	ReviewDecision githubv4.PullRequestReviewDecision
	ReviewRequests struct {
		Nodes []struct {
			AsCodeOwner     githubv4.Boolean
			RequestReviewer struct {
				User struct {
					Login githubv4.String
				} `graphql:"... on User"`
			}
		}
	} `graphql:"reviewRequests(first: 100)"`
	CreatedAt    githubv4.DateTime
	UpdatedAt    githubv4.DateTime
	LastEditedAt githubv4.DateTime
	Labels       struct {
		Nodes []struct {
			Name githubv4.String
		}
	} `graphql:"labels(first: 100)"`
	Assignees struct {
		Nodes []struct {
			Login githubv4.String
		}
	} `graphql:"assignees(first: 100)"`
	Comments struct {
		Nodes []struct {
			Author struct {
				Login githubv4.String
			}
			Body      githubv4.String
			CreatedAt githubv4.DateTime
		}
		PageInfo struct {
			HasNextPage bool
		}
	} `graphql:"comments(first: $limit, orderBy: {direction: DESC, field: UPDATED_AT})"`
}

func (c *Client) FetchTargets(ctx context.Context) (target.Targets, error) {
	targets := target.Targets{}

	var q struct {
		Repogitory struct {
			Issues struct {
				Nodes    []issueNode
				PageInfo struct {
					HasNextPage bool
				}
			} `graphql:"issues(first: $limit, states: OPEN, orderBy: {direction: DESC, field: CREATED_AT})"`
			PullRequests struct {
				Nodes    []pullRequestNode
				PageInfo struct {
					HasNextPage bool
				}
			} `graphql:"pullRequests(first: $limit, states: OPEN, orderBy: {direction: DESC, field: CREATED_AT})"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]interface{}{
		"owner": githubv4.String(c.owner),
		"repo":  githubv4.String(c.repo),
		"limit": githubv4.Int(limit),
	}

	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	if q.Repogitory.Issues.PageInfo.HasNextPage {
		return nil, fmt.Errorf("too many opened issues (limit: %d)", limit)
	}

	if q.Repogitory.PullRequests.PageInfo.HasNextPage {
		return nil, fmt.Errorf("too many opened pull requests (limit: %d)", limit)
	}

	now := time.Now()

	for _, i := range q.Repogitory.Issues.Nodes {
		n := int(i.Number)

		if i.Comments.PageInfo.HasNextPage {
			return nil, fmt.Errorf("too many issue comments (number: %d, limit: %d)", n, limit)
		}
		cc := time.Time{}
		latestComment := struct {
			Author struct {
				Login githubv4.String
			}
			Body      githubv4.String
			CreatedAt githubv4.DateTime
		}{}
		for _, c := range i.Comments.Nodes {
			if cc.Unix() < c.CreatedAt.Unix() {
				latestComment = c
				cc = c.CreatedAt.Time
			}
		}

		labels := []string{}
		for _, l := range i.Labels.Nodes {
			labels = append(labels, string(l.Name))
		}
		assignees := []string{}
		for _, a := range i.Assignees.Nodes {
			assignees = append(assignees, string(a.Login))
		}

		t := &target.Target{
			Number:                   n,
			Title:                    string(i.Title),
			Body:                     string(i.Body),
			URL:                      string(i.URL),
			Author:                   string(i.Author.Login),
			Labels:                   labels,
			Assignees:                assignees,
			IsIssue:                  true,
			IsPullRequest:            false,
			HoursElapsedSinceCreated: int(now.Sub(i.CreatedAt.Time).Hours()),
			HoursElapsedSinceUpdated: int(now.Sub(i.UpdatedAt.Time).Hours()),
			NumberOfComments:         len(i.Comments.Nodes),
			LatestCommentAuthor:      string(latestComment.Author.Login),
		}

		targets[n] = t
	}

	for _, p := range q.Repogitory.PullRequests.Nodes {
		n := int(p.Number)

		if bool(p.IsDraft) {
			// Skip draft pull request
			continue
		}

		if p.Comments.PageInfo.HasNextPage {
			return nil, fmt.Errorf("too many pull request comments (number: %d, limit: %d)", n, limit)
		}
		pc := time.Time{}
		latestComment := struct {
			Author struct {
				Login githubv4.String
			}
			Body      githubv4.String
			CreatedAt githubv4.DateTime
		}{}
		for _, c := range p.Comments.Nodes {
			if pc.Unix() < c.CreatedAt.Unix() {
				latestComment = c
				pc = c.CreatedAt.Time
			}
		}
		isApproved := false
		isReviewRequired := false
		isChangeRequested := false
		switch p.ReviewDecision {
		case githubv4.PullRequestReviewDecisionApproved:
			isApproved = true
		case githubv4.PullRequestReviewDecisionReviewRequired:
			isReviewRequired = true
		case githubv4.PullRequestReviewDecisionChangesRequested:
			isChangeRequested = true
		}
		labels := []string{}
		for _, l := range p.Labels.Nodes {
			labels = append(labels, string(l.Name))
		}
		assignees := []string{}
		for _, a := range p.Assignees.Nodes {
			assignees = append(assignees, string(a.Login))
		}
		reviewers := []string{}
		for _, r := range p.ReviewRequests.Nodes {
			reviewers = append(reviewers, string(r.RequestReviewer.User.Login))
		}

		t := &target.Target{
			Number:                   n,
			Title:                    string(p.Title),
			Body:                     string(p.Body),
			URL:                      string(p.URL),
			Author:                   string(p.Author.Login),
			Labels:                   labels,
			Assignees:                assignees,
			Reviewers:                reviewers,
			IsIssue:                  false,
			IsPullRequest:            true,
			IsApproved:               isApproved,
			IsReviewRequired:         isReviewRequired,
			IsChangeRequested:        isChangeRequested,
			HoursElapsedSinceCreated: int(now.Sub(p.CreatedAt.Time).Hours()),
			HoursElapsedSinceUpdated: int(now.Sub(p.UpdatedAt.Time).Hours()),
			NumberOfComments:         len(p.Comments.Nodes),
			LatestCommentAuthor:      string(latestComment.Author.Login),
		}

		targets[n] = t
	}

	return targets, nil
}

func (c *Client) FetchTarget(ctx context.Context, n int) (*target.Target, error) {
	var q struct {
		Repogitory struct {
			IssueOrPullRequest struct {
				Issue       issueNode       `graphql:"... on Issue"`
				PullRequest pullRequestNode `graphql:"... on PullRequest"`
			} `graphql:"issueOrPullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]interface{}{
		"owner":  githubv4.String(c.owner),
		"repo":   githubv4.String(c.repo),
		"number": githubv4.Int(n),
		"limit":  githubv4.Int(limit),
	}

	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	now := time.Now()

	if strings.Contains(string(q.Repogitory.IssueOrPullRequest.Issue.URL), "/issues/") {
		// Issue
		i := q.Repogitory.IssueOrPullRequest.Issue
		n := int(i.Number)

		if i.Comments.PageInfo.HasNextPage {
			return nil, fmt.Errorf("too many issue comments (number: %d, limit: %d)", n, limit)
		}
		cc := time.Time{}
		latestComment := struct {
			Author struct {
				Login githubv4.String
			}
			Body      githubv4.String
			CreatedAt githubv4.DateTime
		}{}
		for _, c := range i.Comments.Nodes {
			if cc.Unix() < c.CreatedAt.Unix() {
				latestComment = c
				cc = c.CreatedAt.Time
			}
		}

		labels := []string{}
		for _, l := range i.Labels.Nodes {
			labels = append(labels, string(l.Name))
		}
		assignees := []string{}
		for _, a := range i.Assignees.Nodes {
			assignees = append(assignees, string(a.Login))
		}

		t := &target.Target{
			Number:                   n,
			Title:                    string(i.Title),
			Body:                     string(i.Body),
			URL:                      string(i.URL),
			Author:                   string(i.Author.Login),
			Labels:                   labels,
			Assignees:                assignees,
			IsIssue:                  true,
			IsPullRequest:            false,
			HoursElapsedSinceCreated: int(now.Sub(i.CreatedAt.Time).Hours()),
			HoursElapsedSinceUpdated: int(now.Sub(i.UpdatedAt.Time).Hours()),
			NumberOfComments:         len(i.Comments.Nodes),
			LatestCommentAuthor:      string(latestComment.Author.Login),
		}
		return t, nil
	} else {
		// Pull request
		p := q.Repogitory.IssueOrPullRequest.PullRequest
		n := int(p.Number)

		if p.Comments.PageInfo.HasNextPage {
			return nil, fmt.Errorf("too many pull request comments (number: %d, limit: %d)", n, limit)
		}
		pc := time.Time{}
		latestComment := struct {
			Author struct {
				Login githubv4.String
			}
			Body      githubv4.String
			CreatedAt githubv4.DateTime
		}{}
		for _, c := range p.Comments.Nodes {
			if pc.Unix() < c.CreatedAt.Unix() {
				latestComment = c
				pc = c.CreatedAt.Time
			}
		}
		isApproved := false
		isReviewRequired := false
		isChangeRequested := false
		switch p.ReviewDecision {
		case githubv4.PullRequestReviewDecisionApproved:
			isApproved = true
		case githubv4.PullRequestReviewDecisionReviewRequired:
			isReviewRequired = true
		case githubv4.PullRequestReviewDecisionChangesRequested:
			isChangeRequested = true
		}
		labels := []string{}
		for _, l := range p.Labels.Nodes {
			labels = append(labels, string(l.Name))
		}
		assignees := []string{}
		for _, a := range p.Assignees.Nodes {
			assignees = append(assignees, string(a.Login))
		}
		reviewers := []string{}
		for _, r := range p.ReviewRequests.Nodes {
			reviewers = append(reviewers, string(r.RequestReviewer.User.Login))
		}

		t := &target.Target{
			Number:                   n,
			Title:                    string(p.Title),
			Body:                     string(p.Body),
			URL:                      string(p.URL),
			Author:                   string(p.Author.Login),
			Labels:                   labels,
			Assignees:                assignees,
			Reviewers:                reviewers,
			IsIssue:                  false,
			IsPullRequest:            true,
			IsApproved:               isApproved,
			IsReviewRequired:         isReviewRequired,
			IsChangeRequested:        isChangeRequested,
			HoursElapsedSinceCreated: int(now.Sub(p.CreatedAt.Time).Hours()),
			HoursElapsedSinceUpdated: int(now.Sub(p.UpdatedAt.Time).Hours()),
			NumberOfComments:         len(p.Comments.Nodes),
			LatestCommentAuthor:      string(latestComment.Author.Login),
		}

		return t, nil
	}
}

func (c *Client) SetLabels(ctx context.Context, n int, labels []string) error {
	_, _, err := c.v3.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
		Labels: &labels,
	})
	return err
}

func (c *Client) SetAssignees(ctx context.Context, n int, assignees []string) error {
	as := []string{}
	for _, a := range assignees {
		trimed := strings.Trim(a, "@")
		if !strings.Contains(trimed, "/") {
			as = append(as, trimed)
			continue
		}
		splitted := strings.Split(trimed, "/")
		org := splitted[0]
		slug := splitted[1]
		opts := &github.TeamListTeamMembersOptions{}
		users, _, err := c.v3.Teams.ListTeamMembersBySlug(ctx, org, slug, opts)
		if err != nil {
			return err
		}
		for _, u := range users {
			as = append(as, *u.Login)
		}
	}
	as = unique(as)

	if os.Getenv("GITHUB_ASSIGNEES_SAMPLE") != "" {
		sn, err := strconv.Atoi(os.Getenv("GITHUB_ASSIGNEES_SAMPLE"))
		if err != nil {
			return err
		}
		if len(as) < sn {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(as), func(i, j int) { as[i], as[j] = as[j], as[i] })
			as = as[:sn]
		}
	}

	_, _, err := c.v3.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
		Assignees: &as,
	})
	return err
}

func (c *Client) AddComment(ctx context.Context, n int, comment string) error {
	_, _, err := c.v3.Issues.CreateComment(ctx, c.owner, c.repo, n, &github.IssueComment{
		Body: &comment,
	})
	return err
}

func (c *Client) CloseIssue(ctx context.Context, n int) error {
	state := "closed"
	_, _, err := c.v3.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
		State: &state,
	})
	return err
}

func (c *Client) MergePullRequest(ctx context.Context, n int) error {
	_, _, err := c.v3.PullRequests.Merge(ctx, c.owner, c.repo, n, "", &github.PullRequestOptions{})
	return err
}

type roundTripper struct {
	transport   *http.Transport
	accessToken string
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("token %s", rt.accessToken))
	return rt.transport.RoundTrip(r)
}

func httpClient(token string) *http.Client {
	t := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	rt := roundTripper{
		transport:   t,
		accessToken: token,
	}
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: rt,
	}
}

func unique(in []string) []string {
	m := map[string]struct{}{}
	for _, s := range in {
		m[s] = struct{}{}
	}
	u := []string{}
	for s := range m {
		u = append(u, s)
	}
	return u
}
