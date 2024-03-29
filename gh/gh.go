package gh

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/hairyhenderson/go-codeowners"
	"github.com/k1LoW/ghdag/erro"
	"github.com/k1LoW/ghdag/target"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const limit = 100

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
	if v3ep := os.Getenv("GITHUB_API_URL"); v3ep != "" {
		baseEndpoint, err := url.Parse(v3ep)
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

	owner := splitted[0]
	repo := splitted[1]

	_, res, err := v3c.Repositories.Get(ctx, owner, repo)
	scopes := strings.Split(res.Header.Get("X-OAuth-Scopes"), ", ")
	log.Debug().Msg(fmt.Sprintf("the scopes your token has authorized: '%s'", strings.Join(scopes, "', '")))
	if err != nil {
		return nil, err
	}

	return &Client{
		v3:    v3c,
		v4:    v4c,
		owner: owner,
		repo:  repo,
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
	HeadRefName    githubv4.String
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

type pullRequestFilesNode struct {
	Files struct {
		Nodes []struct {
			Path githubv4.String
		}
		PageInfo struct {
			HasNextPage bool
			EndCursor   githubv4.String
		}
	} `graphql:"files(first: $limit, after: $cursor)"`
}

func (c *Client) FetchTargets(ctx context.Context) (target.Targets, error) {
	targets := target.Targets{}

	var q struct {
		Viewer struct {
			Login githubv4.String
		} `graphql:"viewer"`
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
	login := string(q.Viewer.Login)

	for _, i := range q.Repogitory.Issues.Nodes {
		t, err := buildTargetFromIssue(login, i, now)
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
		t, err := c.buildTargetFromPullRequest(ctx, login, p, now)
		if err != nil {
			return nil, err
		}
		targets[t.Number] = t
	}

	return targets, nil
}

func (c *Client) FetchTarget(ctx context.Context, n int) (*target.Target, error) {
	var q struct {
		Viewer struct {
			Login githubv4.String
		} `graphql:"viewer"`
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
	login := string(q.Viewer.Login)

	if strings.Contains(string(q.Repogitory.IssueOrPullRequest.Issue.URL), "/issues/") {
		// Issue
		i := q.Repogitory.IssueOrPullRequest.Issue
		state := strings.ToLower(string(i.State))
		if state != "open" {
			return nil, erro.NewNotOpenError(fmt.Errorf("issue #%d is %s", int(i.Number), state))
		}
		return buildTargetFromIssue(login, i, now)
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
		return c.buildTargetFromPullRequest(ctx, login, p, now)
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
		trimed := strings.TrimPrefix(r, "@")
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
		trimed := strings.TrimPrefix(inu, "@")
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

func buildTargetFromIssue(login string, i issueNode, now time.Time) (*target.Target, error) {
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
		if string(c.Author.Login) != login {
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
		Login:                       login,
	}, nil
}

func (c *Client) buildTargetFromPullRequest(ctx context.Context, login string, p pullRequestNode, now time.Time) (*target.Target, error) {
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
		if string(c.Author.Login) != login {
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
	codeOwnersWhoApproved := []string{}
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
	for _, r := range p.LatestReviews.Nodes {
		u := string(r.Author.Login)
		reviewers = append(reviewers, u)
		if r.State != githubv4.PullRequestReviewStateApproved {
			continue
		}
		reviewersWhoApproved = append(reviewersWhoApproved, u)
	}
	reviewers = unique(reviewers)

	if len(reviewersWhoApproved) > 0 {
		// re-calc code_owners*
		codeOwners = []string{}
		// calcedCodeOwners contains users that exist in the CODEOWNERS file but do not actually exist or do not have permissions.
		calcedCodeOwners, err := c.getCodeOwners(ctx, p)
		if err != nil {
			return nil, err
		}
		for _, u := range reviewersWhoApproved {
			if contains(calcedCodeOwners, u) {
				codeOwnersWhoApproved = append(codeOwnersWhoApproved, u)
			}
		}
		for _, u := range reviewers {
			if contains(calcedCodeOwners, u) {
				codeOwners = append(codeOwners, u)
			}
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
		Login:                       login,
	}, nil
}

func (c *Client) getCodeOwners(ctx context.Context, p pullRequestNode) ([]string, error) {
	// Get CODEOWNERS file
	var cc string
	for _, path := range []string{".github/CODEOWNERS", "docs/CODEOWNERS"} {
		f, _, _, err := c.v3.Repositories.GetContents(ctx, c.owner, c.repo, path, &github.RepositoryContentGetOptions{
			Ref: string(p.HeadRefName),
		})
		if err != nil {
			continue
		}

		switch *f.Encoding {
		case "base64":
			if f.Content == nil {
				break
			}
			c, err := base64.StdEncoding.DecodeString(*f.Content)
			if err != nil {
				break
			}
			cc = string(c)
		case "":
			if f.Content == nil {
				cc = ""
			} else {
				cc = *f.Content
			}
		default:
			break
		}
		break
	}

	if cc == "" {
		return []string{}, nil
	}

	d, err := codeowners.FromReader(strings.NewReader(cc), ".")
	if err != nil {
		return nil, err
	}

	var cursor string
	co := []string{}

	var q struct {
		Repogitory struct {
			PullRequest pullRequestFilesNode `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	for {
		variables := map[string]interface{}{
			"owner":  githubv4.String(c.owner),
			"repo":   githubv4.String(c.repo),
			"number": p.Number,
			"limit":  githubv4.Int(limit),
			"cursor": githubv4.String(cursor),
		}
		if err := c.v4.Query(ctx, &q, variables); err != nil {
			return nil, err
		}
		for _, f := range q.Repogitory.PullRequest.Files.Nodes {
			co = append(co, d.Owners(string(f.Path))...)
		}
		if !q.Repogitory.PullRequest.Files.PageInfo.HasNextPage {
			break
		}
		cursor = string(q.Repogitory.PullRequest.Files.PageInfo.EndCursor)
	}

	codeOwners := []string{}
	for _, o := range unique(co) {
		codeOwners = append(codeOwners, strings.TrimPrefix(o, "@"))
	}
	return codeOwners, nil
}

type GitHubEvent struct {
	Name    string
	Number  int
	State   string
	Payload interface{}
}

func DecodeGitHubEvent() (*GitHubEvent, error) {
	i := &GitHubEvent{}
	n := os.Getenv("GITHUB_EVENT_NAME")
	if n == "" {
		return i, fmt.Errorf("env %s is not set.", "GITHUB_EVENT_NAME")
	}
	i.Name = n
	p := os.Getenv("GITHUB_EVENT_PATH")
	if p == "" {
		return i, fmt.Errorf("env %s is not set.", "GITHUB_EVENT_PATH")
	}
	b, err := ioutil.ReadFile(filepath.Clean(p))
	if err != nil {
		return i, err
	}
	s := struct {
		PullRequest struct {
			Number int    `json:"number,omitempty"`
			State  string `json:"state,omitempty"`
		} `json:"pull_request,omitempty"`
		Issue struct {
			Number int    `json:"number,omitempty"`
			State  string `json:"state,omitempty"`
		} `json:"issue,omitempty"`
	}{}
	if err := json.Unmarshal(b, &s); err != nil {
		return i, err
	}
	switch {
	case s.PullRequest.Number > 0:
		i.Number = s.PullRequest.Number
		i.State = s.PullRequest.State
	case s.Issue.Number > 0:
		i.Number = s.Issue.Number
		i.State = s.Issue.State
	}

	var payload interface{}

	if err := json.Unmarshal(b, &payload); err != nil {
		return i, err
	}

	i.Payload = payload

	return i, nil
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
	u := []string{}
	m := map[string]struct{}{}
	for _, s := range in {
		if _, ok := m[s]; ok {
			continue
		}
		u = append(u, s)
		m[s] = struct{}{}
	}
	return u
}
