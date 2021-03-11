package gh

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/k1LoW/ghdag/erro"
	"github.com/k1LoW/ghdag/target"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const limit = 100
const CommentSigPrefix = "<!-- ghdag:"

type GhClient interface {
	FetchTargets(ctx context.Context) (target.Targets, error)
	FetchTarget(ctx context.Context, n int) (*target.Target, error)
	SetLabels(ctx context.Context, n int, labels []string) error
	SetAssignees(ctx context.Context, n int, assignees []string) error
	SetReviewers(ctx context.Context, n int, reviewers []string) error
	AddComment(ctx context.Context, n int, comment string) error
	CloseIssue(ctx context.Context, n int) error
	MergePullRequest(ctx context.Context, n int) error
	ResolveUsers(ctx context.Context, in []string) ([]string, error)
}

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
	Number    githubv4.Int
	State     githubv4.String
	Title     githubv4.String
	Body      githubv4.String
	URL       githubv4.String
	CreatedAt githubv4.DateTime
	UpdatedAt githubv4.DateTime
	Labels    struct {
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
	State          githubv4.String
	Title          githubv4.String
	Body           githubv4.String
	URL            githubv4.String
	IsDraft        githubv4.Boolean
	ChangedFiles   githubv4.Int
	Mergeable      githubv4.MergeableState
	ReviewDecision githubv4.PullRequestReviewDecision
	ReviewRequests struct {
		Nodes []struct {
			AsCodeOwner       githubv4.Boolean
			RequestedReviewer struct {
				User struct {
					Login githubv4.String
				} `graphql:"... on User"`
				Team struct {
					Organization struct {
						Login githubv4.String
					}
					Slug githubv4.String
				} `graphql:"... on Team"`
			}
		}
	} `graphql:"reviewRequests(first: 100)"`
	LatestReviews struct {
		Nodes []struct {
			Author struct {
				Login githubv4.String
			}
			State githubv4.PullRequestReviewState
		}
	} `graphql:"latestReviews(first: 100)"`
	CreatedAt githubv4.DateTime
	UpdatedAt githubv4.DateTime
	Labels    struct {
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
		t, err := buildTargetFromIssue(i, now)
		if err != nil {
			return nil, err
		}
		targets[t.Number] = t
	}

	for _, p := range q.Repogitory.PullRequests.Nodes {
		if bool(p.IsDraft) {
			// Skip draft pull request
			continue
		}
		t, err := buildTargetFromPullRequest(p, now)
		if err != nil {
			return nil, err
		}
		targets[t.Number] = t
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
		state := strings.ToLower(string(i.State))
		if state != "open" {
			return nil, erro.NewNotOpenError(fmt.Errorf("issue #%d is %s", int(i.Number), state))
		}
		return buildTargetFromIssue(i, now)
	} else {
		// Pull request
		p := q.Repogitory.IssueOrPullRequest.PullRequest
		state := strings.ToLower(string(p.State))
		if state != "open" {
			return nil, erro.NewNotOpenError(fmt.Errorf("pull request #%d is %s", int(p.Number), state))
		}
		if bool(p.IsDraft) {
			return nil, erro.NewNotOpenError(fmt.Errorf("pull request #%d is draft", int(p.Number)))
		}
		return buildTargetFromPullRequest(p, now)
	}
}

func (c *Client) SetLabels(ctx context.Context, n int, labels []string) error {
	_, _, err := c.v3.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
		Labels: &labels,
	})
	return err
}

func (c *Client) SetAssignees(ctx context.Context, n int, assignees []string) error {
	if _, _, err := c.v3.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
		Assignees: &assignees,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetReviewers(ctx context.Context, n int, reviewers []string) error {
	ru := map[string]struct{}{}
	rt := map[string]struct{}{}
	for _, r := range reviewers {
		trimed := strings.Trim(r, "@")
		if strings.Contains(trimed, "/") {
			splitted := strings.Split(trimed, "/")
			rt[splitted[1]] = struct{}{}
			continue
		}
		ru[trimed] = struct{}{}
	}
	current, _, err := c.v3.PullRequests.ListReviewers(ctx, c.owner, c.repo, n, &github.ListOptions{})
	if err != nil {
		return err
	}
	du := []string{}
	dt := []string{}
	for _, u := range current.Users {
		if _, ok := ru[u.GetLogin()]; ok {
			delete(ru, u.GetLogin())
			continue
		}
		du = append(du, u.GetLogin())
	}
	for _, t := range current.Teams {
		if _, ok := rt[t.GetSlug()]; ok {
			delete(rt, t.GetSlug())
			continue
		}
		dt = append(dt, t.GetSlug())
	}
	if len(du) > 0 || len(dt) > 0 {
		if len(du) == 0 {
			du = append(du, "ghdag-dummy")
		}
		if _, err := c.v3.PullRequests.RemoveReviewers(ctx, c.owner, c.repo, n, github.ReviewersRequest{
			Reviewers:     du,
			TeamReviewers: dt,
		}); err != nil {
			return err
		}
	}
	au := []string{}
	at := []string{}
	for k := range ru {
		au = append(au, k)
	}
	for k := range rt {
		at = append(at, k)
	}
	if _, _, err := c.v3.PullRequests.RequestReviewers(ctx, c.owner, c.repo, n, github.ReviewersRequest{
		Reviewers:     au,
		TeamReviewers: at,
	}); err != nil {
		return err
	}
	return nil
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

func (c *Client) ResolveUsers(ctx context.Context, in []string) ([]string, error) {
	res := []string{}
	for _, inu := range in {
		trimed := strings.Trim(inu, "@")
		if !strings.Contains(trimed, "/") {
			res = append(res, trimed)
			continue
		}
		splitted := strings.Split(trimed, "/")
		org := splitted[0]
		slug := splitted[1]
		opts := &github.TeamListTeamMembersOptions{}
		users, _, err := c.v3.Teams.ListTeamMembersBySlug(ctx, org, slug, opts)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			res = append(res, *u.Login)
		}
	}
	return unique(res), nil
}

type roundTripper struct {
	transport   *http.Transport
	accessToken string
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("token %s", rt.accessToken))
	return rt.transport.RoundTrip(r)
}

func buildTargetFromIssue(i issueNode, now time.Time) (*target.Target, error) {
	n := int(i.Number)

	if i.Comments.PageInfo.HasNextPage {
		return nil, fmt.Errorf("too many issue comments (number: %d, limit: %d)", n, limit)
	}
	latestComment := struct {
		Author struct {
			Login githubv4.String
		}
		Body      githubv4.String
		CreatedAt githubv4.DateTime
	}{}
	sort.Slice(i.Comments.Nodes, func(a, b int) bool {
		// CreatedAt DESC
		return (i.Comments.Nodes[a].CreatedAt.Unix() > i.Comments.Nodes[b].CreatedAt.Unix())
	})
	if len(i.Comments.Nodes) > 0 {
		latestComment = i.Comments.Nodes[0]
	}
	numComments := 0
	for _, c := range i.Comments.Nodes {
		if !strings.Contains(string(c.Body), CommentSigPrefix) {
			break
		}
		numComments++
	}

	labels := []string{}
	for _, l := range i.Labels.Nodes {
		labels = append(labels, string(l.Name))
	}
	assignees := []string{}
	for _, a := range i.Assignees.Nodes {
		assignees = append(assignees, string(a.Login))
	}

	return &target.Target{
		Number:                      n,
		State:                       strings.ToLower(string(i.State)),
		Title:                       string(i.Title),
		Body:                        string(i.Body),
		URL:                         string(i.URL),
		Author:                      string(i.Author.Login),
		Labels:                      labels,
		Assignees:                   assignees,
		IsIssue:                     true,
		IsPullRequest:               false,
		HoursElapsedSinceCreated:    int(now.Sub(i.CreatedAt.Time).Hours()),
		HoursElapsedSinceUpdated:    int(now.Sub(i.UpdatedAt.Time).Hours()),
		NumberOfComments:            len(i.Comments.Nodes),
		LatestCommentAuthor:         string(latestComment.Author.Login),
		LatestCommentBody:           string(latestComment.Body),
		NumberOfConsecutiveComments: numComments,
	}, nil
}

func buildTargetFromPullRequest(p pullRequestNode, now time.Time) (*target.Target, error) {
	n := int(p.Number)

	if p.Comments.PageInfo.HasNextPage {
		return nil, fmt.Errorf("too many pull request comments (number: %d, limit: %d)", n, limit)
	}
	latestComment := struct {
		Author struct {
			Login githubv4.String
		}
		Body      githubv4.String
		CreatedAt githubv4.DateTime
	}{}
	sort.Slice(p.Comments.Nodes, func(a, b int) bool {
		// CreatedAt DESC
		return (p.Comments.Nodes[a].CreatedAt.Unix() > p.Comments.Nodes[b].CreatedAt.Unix())
	})
	if len(p.Comments.Nodes) > 0 {
		latestComment = p.Comments.Nodes[0]
	}
	numComments := 0
	for _, c := range p.Comments.Nodes {
		if !strings.Contains(string(c.Body), CommentSigPrefix) {
			break
		}
		numComments++
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
	mergeable := false
	if p.Mergeable == githubv4.MergeableStateMergeable {
		mergeable = true
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
	codeOwners := []string{}
	for _, r := range p.ReviewRequests.Nodes {
		var k string
		if r.RequestedReviewer.User.Login != "" {
			k = string(r.RequestedReviewer.User.Login)
		}
		if r.RequestedReviewer.Team.Slug != "" {
			k = fmt.Sprintf("%s/%s", string(r.RequestedReviewer.Team.Organization.Login), string(r.RequestedReviewer.Team.Slug))
		}
		reviewers = append(reviewers, k)
		if bool(r.AsCodeOwner) {
			codeOwners = append(codeOwners, k)
		}
	}
	reviewersWhoApproved := []string{}
	codeOwnersWhoApproved := []string{}
	for _, r := range p.LatestReviews.Nodes {
		if r.State != githubv4.PullRequestReviewStateApproved {
			continue
		}
		u := string(r.Author.Login)
		reviewersWhoApproved = append(reviewersWhoApproved, u)
		if contains(codeOwners, u) {
			codeOwnersWhoApproved = append(codeOwnersWhoApproved, u)
		}
	}

	return &target.Target{
		Number:                      n,
		State:                       string(p.State),
		Title:                       string(p.Title),
		Body:                        string(p.Body),
		URL:                         string(p.URL),
		Author:                      string(p.Author.Login),
		Labels:                      labels,
		Assignees:                   assignees,
		Reviewers:                   reviewers,
		CodeOwners:                  codeOwners,
		ReviewersWhoApproved:        reviewersWhoApproved,
		CodeOwnersWhoApproved:       codeOwnersWhoApproved,
		IsIssue:                     false,
		IsPullRequest:               true,
		IsApproved:                  isApproved,
		IsReviewRequired:            isReviewRequired,
		IsChangeRequested:           isChangeRequested,
		Mergeable:                   mergeable,
		ChangedFiles:                int(p.ChangedFiles),
		HoursElapsedSinceCreated:    int(now.Sub(p.CreatedAt.Time).Hours()),
		HoursElapsedSinceUpdated:    int(now.Sub(p.UpdatedAt.Time).Hours()),
		NumberOfComments:            len(p.Comments.Nodes),
		LatestCommentAuthor:         string(latestComment.Author.Login),
		LatestCommentBody:           string(latestComment.Body),
		NumberOfConsecutiveComments: numComments,
	}, nil
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

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
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
