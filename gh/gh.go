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
)

type Client struct {
	client *github.Client
	owner  string
	repo   string
}

// NewClient return Client
func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_TOKEN")
	}
	c := github.NewClient(httpClient(token))
	if baseURL := os.Getenv("GITHUB_API_URL"); baseURL != "" {
		baseEndpoint, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(baseEndpoint.Path, "/") {
			baseEndpoint.Path += "/"
		}
		c.BaseURL = baseEndpoint
	}
	ownerrepo := os.Getenv("GITHUB_REPOSITORY")
	if ownerrepo == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_REPOSITORY")
	}
	splitted := strings.Split(ownerrepo, "/")

	return &Client{
		client: c,
		owner:  splitted[0],
		repo:   splitted[1],
	}, nil
}

// Target is Issue or Pull request
type Target struct {
	i *github.Issue
}

func (t *Target) Dump() map[string]interface{} {
	return map[string]interface{}{
		"number":                      t.Number(),
		"title":                       t.Title(),
		"body":                        t.Body(),
		"labels":                      t.Labels(),
		"is_issue":                    t.IsIssue(),
		"is_pull_request":             t.IsPullRequest(),
		"hours_elapsed_since_created": t.HoursElapsedSinceCreated(),
		"hours_elapsed_since_updated": t.HoursElapsedSinceUpdated(),
	}
}

func (t *Target) Number() int {
	return t.i.GetNumber()
}

func (t *Target) Title() string {
	return t.i.GetTitle()
}

func (t *Target) Body() string {
	return t.i.GetBody()
}

func (t *Target) URL() string {
	return t.i.GetHTMLURL()
}

func (t *Target) Labels() []string {
	labels := []string{}
	for _, l := range t.i.Labels {
		labels = append(labels, *l.Name)
	}
	return labels
}

func (t *Target) IsIssue() bool {
	return !t.i.IsPullRequest()
}

func (t *Target) IsPullRequest() bool {
	return t.i.IsPullRequest()
}

func (t *Target) HoursElapsedSinceCreated() int {
	now := time.Now()
	d := now.Sub(t.i.GetCreatedAt())
	return int(d.Hours())
}

func (t *Target) HoursElapsedSinceUpdated() int {
	now := time.Now()
	d := now.Sub(t.i.GetUpdatedAt())
	return int(d.Hours())
}

type Targets map[int]*Target

func (targets Targets) MaxDigits() int {
	digits := 0
	for _, t := range targets {
		if digits < len(fmt.Sprintf("%d", t.Number())) {
			digits = len(fmt.Sprintf("%d", t.Number()))
		}
	}
	return digits
}

func NewTarget(i *github.Issue) *Target {
	return &Target{
		i: i,
	}
}

func (c *Client) FetchTargets(ctx context.Context) (Targets, error) {
	issues, _, err := c.client.Issues.ListByRepo(ctx, c.owner, c.repo, &github.IssueListByRepoOptions{
		State: "open",
	})
	if err != nil {
		return nil, err
	}
	targets := Targets{}
	for _, i := range issues {
		targets[*i.Number] = NewTarget(i)
	}
	return targets, nil
}

func (c *Client) FetchTarget(ctx context.Context, n int) (*Target, error) {
	i, _, err := c.client.Issues.Get(ctx, c.owner, c.repo, n)
	if err != nil {
		return nil, err
	}
	return NewTarget(i), nil
}

func (c *Client) SetLabels(ctx context.Context, n int, labels []string) error {
	_, _, err := c.client.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
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
		users, _, err := c.client.Teams.ListTeamMembersBySlug(ctx, org, slug, opts)
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

	_, _, err := c.client.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
		Assignees: &as,
	})
	return err
}

func (c *Client) AddComment(ctx context.Context, n int, comment string) error {
	_, _, err := c.client.Issues.CreateComment(ctx, c.owner, c.repo, n, &github.IssueComment{
		Body: &comment,
	})
	return err
}

func (c *Client) CloseIssue(ctx context.Context, n int) error {
	state := "closed"
	_, _, err := c.client.Issues.Edit(ctx, c.owner, c.repo, n, &github.IssueRequest{
		State: &state,
	})
	return err
}

func (c *Client) MergePullRequest(ctx context.Context, n int) error {
	_, _, err := c.client.PullRequests.Merge(ctx, c.owner, c.repo, n, "", &github.PullRequestOptions{})
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
