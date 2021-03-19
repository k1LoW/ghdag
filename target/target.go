package target

import (
	"fmt"

	"github.com/goccy/go-json"
)

// Target is Issue or Pull request
type Target struct {
	Number                      int      `json:"number"`
	State                       string   `json:"state"`
	Title                       string   `json:"title"`
	Body                        string   `json:"body"`
	URL                         string   `json:"url"`
	Author                      string   `json:"author"`
	Labels                      []string `json:"labels"`
	Assignees                   []string `json:"assignees"`
	Reviewers                   []string `json:"reviewers"`
	CodeOwners                  []string `json:"code_owners"`
	ReviewersWhoApproved        []string `json:"reviewers_who_approved"`
	CodeOwnersWhoApproved       []string `json:"code_owners_who_approved"`
	IsIssue                     bool     `json:"is_issue"`
	IsPullRequest               bool     `json:"is_pull_request"`
	IsApproved                  bool     `json:"is_approved"`
	IsReviewRequired            bool     `json:"is_review_required"`
	IsChangeRequested           bool     `json:"is_change_requested"`
	Mergeable                   bool     `json:"mergeable"`
	ChangedFiles                int      `json:"changed_files"`
	HoursElapsedSinceCreated    int      `json:"hours_elapsed_since_created"`
	HoursElapsedSinceUpdated    int      `json:"hours_elapsed_since_updated"`
	NumberOfComments            int      `json:"number_of_comments"`
	LatestCommentAuthor         string   `json:"latest_comment_author"`
	LatestCommentBody           string   `json:"latest_comment_body"`
	NumberOfConsecutiveComments int      `json:"-"`

	Login string `json:"login"`
}

func (t *Target) NoCodeOwnerReviewers() []string {
	nr := []string{}
	for _, r := range t.Reviewers {
		if contains(t.CodeOwners, r) {
			continue
		}
		nr = append(nr, r)
	}
	return nr
}

func (t *Target) Dump() map[string]interface{} {
	b, _ := json.Marshal(t)
	v := map[string]interface{}{}
	_ = json.Unmarshal(b, &v)
	return v
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

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
