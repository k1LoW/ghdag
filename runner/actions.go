package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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
	sortStringSlice(i.Labels)
	sortStringSlice(labels)
	r.log(fmt.Sprintf("Set labels: %s", strings.Join(labels, ", ")))
	if cmp.Equal(i.Labels, labels) {
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(labels, ", ")))
	}
	return r.github.SetLabels(ctx, i.Number, labels)
}

func (r *Runner) PerformAssigneesAction(ctx context.Context, i *target.Target, assignees []string) error {
	as, err := r.github.ResolveUsers(ctx, assignees)
	if err != nil {
		return err
	}
	sortStringSlice(i.Assignees)
	as, err = r.sampleByEnv(as, "GITHUB_ASSIGNEES_SAMPLE")
	if err != nil {
		return err
	}
	sortStringSlice(as)
	r.log(fmt.Sprintf("Set assignees: %s", strings.Join(as, ", ")))
	if cmp.Equal(i.Assignees, as) {
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(as, ", ")))
	}
	return r.github.SetAssignees(ctx, i.Number, as)
}

func (r *Runner) PerformReviewersAction(ctx context.Context, i *target.Target, reviewers []string) error {
	if contains(reviewers, i.Author) {
		r.debuglog(fmt.Sprintf("Exclude author from reviewers: %s", reviewers))
		reviewers = exclude(reviewers, i.Author)
	}
	reviewers, err := r.sampleByEnv(reviewers, "GITHUB_REVIEWERS_SAMPLE")
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
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(reviewers, ", ")))
	}
	return r.github.SetReviewers(ctx, i.Number, ra)
}

func (r *Runner) PerformCommentAction(ctx context.Context, i *target.Target, comment, sig string) error {
	c, err := env.ParseWithEnviron(comment, env.EnvMap())
	if err != nil {
		return err
	}
	mentions, err := env.ToSlice(os.Getenv("GITHUB_COMMENT_MENTIONS"))
	if err != nil {
		return err
	}
	mentions, err = r.sampleByEnv(mentions, "GITHUB_COMMENT_MENTIONS_SAMPLE")
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Add comment: %s", c))
	if i.NumberOfConsecutiveComments >= 5 {
		return fmt.Errorf("Too many comments in a row by ghdag: %d", i.NumberOfConsecutiveComments)
	}

	c = fmt.Sprintf("%s\n%s\n", c, sig)

	if i.LatestCommentBody == c {
		return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", c))
	}
	return r.github.AddComment(ctx, i.Number, c, mentions)
}

func (r *Runner) PerformStateAction(ctx context.Context, i *target.Target, state string) error {
	r.log(fmt.Sprintf("Change state: %s", state))
	switch state {
	case "close", "closed":
		return r.github.CloseIssue(ctx, i.Number)
	case "merge", "merged":
		return r.github.MergePullRequest(ctx, i.Number)
	default:
		return fmt.Errorf("invalid state: %s", state)
	}
}

func (r *Runner) PerformNotifyAction(ctx context.Context, i *target.Target, notify string) error {
	n, err := env.ParseWithEnviron(notify, env.EnvMap())
	if err != nil {
		return err
	}
	mentions, err := env.ToSlice(os.Getenv("SLACK_MENTIONS"))
	if err != nil {
		return err
	}
	mentions, err = r.sampleByEnv(mentions, "SLACK_MENTIONS_SAMPLE")
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Send notification: %s", n))
	return r.slack.PostMessage(ctx, n, mentions)
}

func (r *Runner) performNextAction(ctx context.Context, i *target.Target, t *task.Task, q chan TaskQueue, next []string) error {
	r.log(fmt.Sprintf("Call next task: %s", strings.Join(next, ", ")))
	for _, id := range next {
		nt, err := r.config.Tasks.Find(id)
		if err != nil {
			return err
		}
		q <- TaskQueue{
			target:     i,
			task:       nt,
			called:     true,
			callerTask: t,
			callerSeed: r.seed,
		}
	}
	return nil
}
