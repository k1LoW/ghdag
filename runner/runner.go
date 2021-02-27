package runner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/antonmedv/expr"
	"github.com/k1LoW/exec"
	"github.com/k1LoW/ghdag/config"
	"github.com/k1LoW/ghdag/env"
	"github.com/k1LoW/ghdag/gh"
	"github.com/k1LoW/ghdag/slk"
	"github.com/k1LoW/ghdag/target"
	"github.com/k1LoW/ghdag/task"
	"github.com/rs/zerolog/log"
)

type Runner struct {
	config    *config.Config
	github    *gh.Client
	slack     *slk.Client
	envCache  []string
	mu        sync.Mutex
	logPrefix string
}

func New(c *config.Config) (*Runner, error) {
	return &Runner{
		config:    c,
		github:    nil,
		slack:     nil,
		envCache:  os.Environ(),
		logPrefix: "",
	}, nil
}

type TaskQueue struct {
	target *target.Target
	task   *task.Task
	called bool
}

func (r *Runner) Run(ctx context.Context) error {
	r.logPrefix = ""
	r.log("Start session")
	defer func() {
		_ = r.revertEnv()
	}()
	if err := r.config.Env.Setenv(); err != nil {
		return err
	}
	gc, err := gh.NewClient()
	if err != nil {
		return err
	}
	r.github = gc
	sc, err := slk.NewClient()
	if err != nil {
		return err
	}
	r.slack = sc

	r.log(fmt.Sprintf("Fetch open issues and pull requests from %s", os.Getenv("GITHUB_REPOSITORY")))
	targets, err := r.github.FetchTargets(ctx)
	if err != nil {
		return err
	}
	maxDigits := targets.MaxDigits()
	r.log(fmt.Sprintf("%d issues and pull requests are fetched", len(targets)))
	tasks := r.config.Tasks
	r.log(fmt.Sprintf("%d tasks are loaded", len(tasks)))
	maxLength := tasks.MaxLengthID()

	q := make(chan TaskQueue, len(tasks)*len(targets))
	for _, i := range targets {
		for _, t := range tasks {
			q <- TaskQueue{
				target: i,
				task:   t,
			}
		}
	}

L:
	for {
		if len(q) == 0 {
			close(q)
		}

		tq, ok := <-q
		if !ok {
			break
		}

		n := tq.target.Number
		id := tq.task.Id
		r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] ", maxDigits, maxLength), n, id)

		if tq.task.If != "" {
			do, err := expr.Eval(fmt.Sprintf("(%s) == true", tq.task.If), tq.target.Dump())
			if err != nil {
				r.errlog(fmt.Sprintf("%s", err))
				continue L
			}
			if !do.(bool) {
				r.debuglog(fmt.Sprintf("Skip: %s", tq.task.If))
				continue L
			}
		}
		if tq.task.If == "" && !tq.called {
			r.debuglog(fmt.Sprintf("Skip: %s", "(non `if:` section)"))
			continue L
		}

		if tq.called {
			// Update target
			target, err := r.github.FetchTarget(ctx, tq.target.Number)
			if err != nil {
				return err
			}
			tq.target = target
		}

		r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [DO] ", maxDigits, maxLength), n, id)
		if err := r.perform(ctx, tq.task.Do, tq.target, tq.task, q); err == nil {
			r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [OK] ", maxDigits, maxLength), n, id)
			if err := r.perform(ctx, tq.task.Ok, tq.target, tq.task, q); err != nil {
				r.errlog(fmt.Sprintf("%s", err))
				continue L
			}
		} else {
			r.errlog(fmt.Sprintf("%s", err))
			r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [NG] ", maxDigits, maxLength), n, id)
			if err := r.perform(ctx, tq.task.Ng, tq.target, tq.task, q); err != nil {
				r.errlog(fmt.Sprintf("%s", err))
				continue L
			}
		}
	}

	r.logPrefix = ""
	r.log("Session finished")
	return nil
}

func (r *Runner) perform(ctx context.Context, a *task.Action, i *target.Target, t *task.Task, q chan TaskQueue) error {
	if a == nil {
		return nil
	}

	r.mu.Lock()
	defer func() {
		_ = r.revertEnv()
		r.mu.Unlock()
	}()

	if err := os.Setenv("GHDAG_TARGET_NUMBER", fmt.Sprintf("%d", i.Number)); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_TARGET_URL", i.URL); err != nil {
		return err
	}
	if err := os.Setenv("GHDAG_TASK_ID", t.Id); err != nil {
		return err
	}
	if err := r.config.Env.Setenv(); err != nil {
		return err
	}
	if err := t.Env.Setenv(); err != nil {
		return err
	}

	switch {
	case a.Run != "":
		c := exec.CommandContext(ctx, "sh", "-c", a.Run)
		c.Env = os.Environ()
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		return nil
	case len(a.Labels) > 0:
		r.log(fmt.Sprintf("Set labels: %s", strings.Join(a.Labels, ", ")))
		return r.github.SetLabels(ctx, i.Number, a.Labels)
	case len(a.Assignees) > 0:
		r.log(fmt.Sprintf("Set assignees: %s", strings.Join(a.Assignees, ", ")))
		return r.github.SetAssignees(ctx, i.Number, a.Assignees)
	case len(a.Reviewers) > 0:
		r.log(fmt.Sprintf("Set reviewers: %s", strings.Join(a.Reviewers, ", ")))
		return r.github.SetReviewers(ctx, i.Number, a.Reviewers)
	case a.Comment != "":
		r.log(fmt.Sprintf("Add comment: %s", a.Comment))
		return r.github.AddComment(ctx, i.Number, a.Comment)
	case a.State != "":
		r.log(fmt.Sprintf("Change state: %s", a.State))
		switch a.State {
		case "close", "closed":
			return r.github.CloseIssue(ctx, i.Number)
		case "merge", "merged":
			return r.github.MergePullRequest(ctx, i.Number)
		default:
			return fmt.Errorf("invalid state: %s", a.State)
		}
	case a.Notify != "":
		r.log(fmt.Sprintf("Send notification: %s", a.Notify))
		return r.slack.PostMessage(ctx, a.Notify)
	case len(a.Next) > 0:
		r.log(fmt.Sprintf("Call next task: %s", strings.Join(a.Next, ", ")))
		for _, id := range a.Next {
			t, err := r.config.Tasks.Find(id)
			if err != nil {
				return err
			}
			q <- TaskQueue{
				target: i,
				task:   t,
				called: true,
			}
		}
	}
	return nil
}

func (r *Runner) log(m string) {
	log.Info().Msg(fmt.Sprintf("%s%s", r.logPrefix, m))
}

func (r *Runner) errlog(m string) {
	log.Error().Msg(fmt.Sprintf("%s%s", r.logPrefix, m))
}

func (r *Runner) debuglog(m string) {
	log.Debug().Msg(fmt.Sprintf("%s%s", r.logPrefix, m))
}

func (r *Runner) revertEnv() error {
	return env.Revert(r.envCache)
}
