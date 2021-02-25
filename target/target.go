package target

import "fmt"

// Target is Issue or Pull request
type Target struct {
	Number                   int
	Title                    string
	Body                     string
	URL                      string
	Labels                   []string
	Assignees                []string
	IsIssue                  bool
	IsPullRequest            bool
	HoursElapsedSinceCreated int
	HoursElapsedSinceUpdated int
}

func (t *Target) Dump() map[string]interface{} {
	return map[string]interface{}{
		"number":                      t.Number,
		"title":                       t.Title,
		"body":                        t.Body,
		"labels":                      t.Labels,
		"assignees":                   t.Assignees,
		"is_issue":                    t.IsIssue,
		"is_pull_request":             t.IsPullRequest,
		"hours_elapsed_since_created": t.HoursElapsedSinceCreated,
		"hours_elapsed_since_updated": t.HoursElapsedSinceUpdated,
	}
}

type Targets map[int]*Target

func (targets Targets) MaxDigits() int {
	digits := 0
	for _, t := range targets {
		if digits < len(fmt.Sprintf("%d", t.Number)) {
			digits = len(fmt.Sprintf("%d", t.Number))
		}
	}
	return digits
}
