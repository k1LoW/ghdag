package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/exec"
	"github.com/k1LoW/ghdag/env"
	"github.com/k1LoW/ghdag/erro"
	"github.com/k1LoW/ghdag/target"
	"github.com/k1LoW/ghdag/task"
)

func (r *Runner) PerformRunAction(ctx context.Context, _ *target.Target, command string) error {
	c := exec.CommandContext(ctx, "sh", "-c", command)
	c.Env = os.Environ()
	outbuf := new(bytes.Buffer)
	outmr := io.MultiWriter(os.Stdout, outbuf)
	c.Stdout = outmr
	errbuf := new(bytes.Buffer)
	errmr := io.MultiWriter(os.Stderr, errbuf)
	c.Stderr = errmr
	if err := c.Run(); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_ACTION_RUN_STDOUT", outbuf.String()); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_ACTION_RUN_STDERR", errbuf.String()); err != nil {
		return err
	}
	return nil
}

func (r *Runner) PerformLabelsAction(ctx context.Context, i *target.Target, labels []string) error {
	b := os.Getenv("GHDAG_ACTION_LABELS_BEHAVIOR")
	switch b {
	case "add":
		r.log(fmt.Sprintf("Add labels: %s", strings.Join(labels, ", ")))
		labels = unique(append(labels, i.Labels...))
	case "remove":
		r.log(fmt.Sprintf("Remove labels: %s", strings.Join(labels, ", ")))
		removed := []string{}
		for _, l := range i.Labels {
			if contains(labels, l) {
				continue
			}
			removed = append(removed, l)
		}
		labels = removed
	case "replace", "":
		r.log(fmt.Sprintf("Replace labels: %s", strings.Join(labels, ", ")))
	default:
		return fmt.Errorf("invalid behavior: %s", b)
	}

	sortStringSlice(i.Labels)
	sortStringSlice(labels)
	if cmp.Equal(i.Labels, labels) {
		if err := os.Setenv("GHDAG_ACTION_LABELS_UPDATED", env.Join(labels)); err != nil {
			return err
		}
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(labels, ", ")))
	}
	if err := r.github.SetLabels(ctx, i.Number, labels); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_ACTION_LABELS_UPDATED", env.Join(labels)); err != nil {
		return err
	}
	return nil
}

func (r *Runner) PerformAssigneesAction(ctx context.Context, i *target.Target, assignees []string) error {
	assignees = r.config.LinkedNames.ToGithubNames(assignees)
	as, err := r.github.ResolveUsers(ctx, assignees)
	if err != nil {
		return err
	}
	sortStringSlice(i.Assignees)
	as, err = r.sample(as, "GITHUB_ASSIGNEES_SAMPLE")
	if err != nil {
		return err
	}
	sortStringSlice(as)
	r.log(fmt.Sprintf("Set assignees: %s", strings.Join(as, ", ")))
	if cmp.Equal(i.Assignees, as) {
		if err := os.Setenv("GHDAG_ACTION_ASSIGNEES_UPDATED", env.Join(as)); err != nil {
			return err
		}
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(as, ", ")))
	}
	if err := r.github.SetAssignees(ctx, i.Number, as); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_ACTION_ASSIGNEES_UPDATED", env.Join(as)); err != nil {
		return err
	}
	return nil
}

func (r *Runner) PerformReviewersAction(ctx context.Context, i *target.Target, reviewers []string) error {
	reviewers = r.config.LinkedNames.ToGithubNames(reviewers)
	if contains(reviewers, i.Author) {
		r.debuglog(fmt.Sprintf("Exclude author from reviewers: %s", reviewers))
		if err := r.setExcludeKey(reviewers, i.Author); err != nil {
			return err
		}
	}
	reviewers, err := r.sample(reviewers, "GITHUB_REVIEWERS_SAMPLE")
	if err != nil {
		return err
	}
	if len(reviewers) == 0 {
		return erro.NewNoReviewerError(errors.New("no reviewers to assign"))
	}

	r.log(fmt.Sprintf("Set reviewers: %s", strings.Join(reviewers, ", ")))

	rb := i.NoCodeOwnerReviewers()
	sortStringSlice(rb)

	ra := []string{}
	for _, r := range reviewers {
		if contains(i.CodeOwners, r) {
			continue
		}
		ra = append(ra, r)
	}
	sortStringSlice(ra)

	if len(ra) == 0 || cmp.Equal(rb, ra) {
		if err := os.Setenv("GHDAG_ACTION_REVIEWERS_UPDATED", env.Join(ra)); err != nil {
			return err
		}
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(reviewers, ", ")))
	}
	if err := r.github.SetReviewers(ctx, i.Number, ra); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_ACTION_REVIEWERS_UPDATED", env.Join(ra)); err != nil {
		return err
	}
	return nil
}

func (r *Runner) PerformCommentAction(ctx context.Context, i *target.Target, comment string) error {
	c := os.ExpandEnv(comment)
	mentions, err := env.Split(os.Getenv("GITHUB_COMMENT_MENTIONS"))
	if err != nil {
		return err
	}
	mentions, err = r.sample(mentions, "GITHUB_COMMENT_MENTIONS_SAMPLE")
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Add comment: %s", c))

	max, err := strconv.Atoi(os.Getenv("GHDAG_ACTION_COMMENT_MAX"))
	if err != nil {
		max = 5
	}

	if i.NumberOfConsecutiveComments >= max {
		return fmt.Errorf("Too many comments in a row by same login: %d", i.NumberOfConsecutiveComments)
	}

	if i.LatestCommentBody == c {
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", c))
	}

	fm := []string{}
	for _, m := range mentions {
		if !strings.HasPrefix(m, "@") {
			m = fmt.Sprintf("@%s", m)
		}
		fm = append(fm, m)
	}
	if len(fm) > 0 {
		c = fmt.Sprintf("%s %s", strings.Join(fm, " "), c)
	}
	if err := r.github.AddComment(ctx, i.Number, c); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_ACTION_COMMENT_CREATED", c); err != nil {
		return err
	}
	return nil
}

func (r *Runner) PerformStateAction(ctx context.Context, i *target.Target, state string) error {
	r.log(fmt.Sprintf("Change state: %s", state))
	switch state {
	case "close", "closed":
		if err := r.github.CloseIssue(ctx, i.Number); err != nil {
			return err
		}
		state = "closed"
	case "merge", "merged":
		if err := r.github.MergePullRequest(ctx, i.Number); err != nil {
			return err
		}
		state = "merged"
	default:
		return fmt.Errorf("invalid state: %s", state)
	}
	if err := os.Setenv("GHDAG_ACTION_STATE_CHANGED", state); err != nil {
		return err
	}
	return nil
}

func (r *Runner) PerformNotifyAction(ctx context.Context, _ *target.Target, notify string) error {
	n := os.ExpandEnv(notify)
	mentions, err := env.Split(os.Getenv("SLACK_MENTIONS"))
	if err != nil {
		return err
	}
	mentions = r.config.LinkedNames.ToSlackNames(mentions)
	mentions, err = r.sample(mentions, "SLACK_MENTIONS_SAMPLE")
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Send notification: %s", n))
	if os.Getenv("SLACK_WEBHOOK_URL") != "" && len(mentions) > 0 {
		return errors.New("notification using webhook does not support mentions")
	}
	links := []string{}
	for _, m := range mentions {
		l, err := r.slack.GetMentionLinkByName(ctx, m)
		if err != nil {
			return err
		}
		links = append(links, l)
	}
	if len(links) > 0 {
		n = fmt.Sprintf("%s %s", strings.Join(links, " "), n)
	}
	if err := r.slack.PostMessage(ctx, n); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_ACTION_NOTIFY_SENT", n); err != nil {
		return err
	}
	return nil
}

var propagatableEnv = []string{
	"GHDAG_ACTION_RUN_STDOUT",
	"GHDAG_ACTION_RUN_STDERR",
	"GHDAG_ACTION_LABELS_UPDATED",
	"GHDAG_ACTION_ASSIGNEES_UPDATED",
	"GHDAG_ACTION_REVIEWERS_UPDATED",
	"GHDAG_ACTION_COMMENT_CREATED",
	"GHDAG_ACTION_STATE_CHANGED",
	"GHDAG_ACTION_NOTIFY_SENT",
	"GHDAG_ACTION_DO_ERROR",
}

func (r *Runner) performNextAction(ctx context.Context, i *target.Target, t *task.Task, q chan TaskQueue, next []string) error {
	r.log(fmt.Sprintf("Call next task: %s", strings.Join(next, ", ")))

	callerEnv := env.Env{}
	for _, k := range propagatableEnv {
		if v, ok := os.LookupEnv(k); ok {
			callerEnv[k] = v
		}
	}

	for _, id := range next {
		nt, err := r.config.Tasks.Find(id)
		if err != nil {
			return err
		}
		q <- TaskQueue{
			target:           i,
			task:             nt,
			called:           true,
			callerTask:       t,
			callerSeed:       r.seed,
			callerExcludeKey: r.excludeKey,
			callerEnv:        callerEnv,
		}
	}
	return nil
}
