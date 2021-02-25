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

func (c *Client) FetchTargets(ctx context.Context) (target.Targets, error) {
	targets := target.Targets{}

	limit := 99

	var q struct {
		Repogitory struct {
			Issues struct {
				Nodes []struct {
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
					} `graphql:"comments(last: 1)"`
				}
			} `graphql:"issues(last: $limit, states: OPEN)"`
			PullRequests struct {
				Nodes []struct {
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
					} `graphql:"comments(last: 1)"`
				}
			} `graphql:"pullRequests(last: $limit, states: OPEN)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]interface{}{
		"owner": githubv4.String(c.owner),
		"repo":  githubv4.String(c.repo),
		"limit": githubv4.Int(limit + 1),
	}

	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	if len(q.Repogitory.Issues.Nodes) > limit {
		return nil, fmt.Errorf("too many opened issues (limit :%d)", limit)
	}

	if len(q.Repogitory.PullRequests.Nodes) > limit {
		return nil, fmt.Errorf("too many opened pull requests (limit :%d)", limit)
	}

	now := time.Now()

	for _, i := range q.Repogitory.Issues.Nodes {
		labels := []string{}
		for _, l := range i.Labels.Nodes {
			labels = append(labels, string(l.Name))
		}
		assignees := []string{}
		for _, a := range i.Assignees.Nodes {
			assignees = append(assignees, string(a.Login))
		}

		n := int(i.Number)
		t := &target.Target{
			Number:                   n,
			Title:                    string(i.Title),
			Body:                     string(i.Body),
			URL:                      string(i.URL),
			Labels:                   labels,
			Assignees:                assignees,
			IsIssue:                  true,
			IsPullRequest:            false,
			HoursElapsedSinceCreated: int(now.Sub(i.CreatedAt.Time).Hours()),
			HoursElapsedSinceUpdated: int(now.Sub(i.UpdatedAt.Time).Hours()),
		}

		targets[n] = t
	}

	for _, p := range q.Repogitory.PullRequests.Nodes {
		labels := []string{}
		for _, l := range p.Labels.Nodes {
			labels = append(labels, string(l.Name))
		}
		assignees := []string{}
		for _, a := range p.Assignees.Nodes {
			assignees = append(assignees, string(a.Login))
		}

		n := int(p.Number)
		t := &target.Target{
			Number:                   n,
			Title:                    string(p.Title),
			Body:                     string(p.Body),
			URL:                      string(p.URL),
			Labels:                   labels,
			Assignees:                assignees,
			IsIssue:                  false,
			IsPullRequest:            true,
			HoursElapsedSinceCreated: int(now.Sub(p.CreatedAt.Time).Hours()),
			HoursElapsedSinceUpdated: int(now.Sub(p.UpdatedAt.Time).Hours()),
		}

		targets[n] = t
	}

	return targets, nil
}

func (c *Client) FetchTarget(ctx context.Context, n int) (*target.Target, error) {
	i, _, err := c.v3.Issues.Get(ctx, c.owner, c.repo, n)
	if err != nil {
		return nil, err
	}
	if i.IsPullRequest() {
		return c.fetchPullRequest(ctx, n)
	}
	return c.fetchIssue(ctx, n)
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

func (c *Client) fetchIssue(ctx context.Context, n int) (*target.Target, error) {
	var q struct {
		Repogitory struct {
			Issue struct {
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
				} `graphql:"comments(last: 1)"`
			} `graphql:"issue(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]interface{}{
		"owner":  githubv4.String(c.owner),
		"repo":   githubv4.String(c.repo),
		"number": githubv4.Int(n),
	}

	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	now := time.Now()

	i := q.Repogitory.Issue
	labels := []string{}
	for _, l := range i.Labels.Nodes {
		labels = append(labels, string(l.Name))
	}
	assignees := []string{}
	for _, a := range i.Assignees.Nodes {
		assignees = append(assignees, string(a.Login))
	}

	t := &target.Target{
		Number:                   int(i.Number),
		Title:                    string(i.Title),
		Body:                     string(i.Body),
		URL:                      string(i.URL),
		Labels:                   labels,
		Assignees:                assignees,
		IsIssue:                  true,
		IsPullRequest:            false,
		HoursElapsedSinceCreated: int(now.Sub(i.CreatedAt.Time).Hours()),
		HoursElapsedSinceUpdated: int(now.Sub(i.UpdatedAt.Time).Hours()),
	}

	return t, nil
}

func (c *Client) fetchPullRequest(ctx context.Context, n int) (*target.Target, error) {
	var q struct {
		Repogitory struct {
			PullRequest struct {
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
				} `graphql:"comments(last: 1)"`
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]interface{}{
		"owner":  githubv4.String(c.owner),
		"repo":   githubv4.String(c.repo),
		"number": githubv4.Int(n),
	}

	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	now := time.Now()

	p := q.Repogitory.PullRequest
	labels := []string{}
	for _, l := range p.Labels.Nodes {
		labels = append(labels, string(l.Name))
	}
	assignees := []string{}
	for _, a := range p.Assignees.Nodes {
		assignees = append(assignees, string(a.Login))
	}

	t := &target.Target{
		Number:                   int(p.Number),
		Title:                    string(p.Title),
		Body:                     string(p.Body),
		URL:                      string(p.URL),
		Labels:                   labels,
		Assignees:                assignees,
		IsIssue:                  false,
		IsPullRequest:            true,
		HoursElapsedSinceCreated: int(now.Sub(p.CreatedAt.Time).Hours()),
		HoursElapsedSinceUpdated: int(now.Sub(p.UpdatedAt.Time).Hours()),
	}

	return t, nil
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
