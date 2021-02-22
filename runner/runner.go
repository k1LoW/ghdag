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
	"github.com/k1LoW/ghdag/gh"
	"github.com/k1LoW/ghdag/slk"
	"github.com/k1LoW/ghdag/task"
	"github.com/rs/zerolog/log"
)

type Runner struct {
	config    *config.Config
	github    *gh.Client
	slack     *slk.Client
	env       []string
	mu        sync.Mutex
	logPrefix string
}

func New(c *config.Config) (*Runner, error) {
	gc, err := gh.NewClient()
	if err != nil {
		return nil, err
	}
	sc, err := slk.NewClient()
	if err != nil {
		return nil, err
	}
	return &Runner{
		config:    c,
		github:    gc,
		slack:     sc,
		env:       os.Environ(),
		logPrefix: "",
	}, nil
}

type TaskQueue struct {
	target *gh.Target
	task   *task.Task
	force  bool
}

func (r *Runner) Run(ctx context.Context) error {
	r.logPrefix = ""
	r.log("Start session")
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

		n := tq.target.Number()
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
		if tq.task.If == "" && !tq.force {
			r.debuglog(fmt.Sprintf("Skip: %s", "(non `if:` section)"))
			continue L
		}

		r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds][DO] ", maxDigits, maxLength), n, id)
		if err := r.Perform(ctx, tq.task.Do, tq.target, tq.task, q); err == nil {
			r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds][OK] ", maxDigits, maxLength), n, id)
			if err := r.Perform(ctx, tq.task.Ok, tq.target, tq.task, q); err != nil {
				r.errlog(fmt.Sprintf("%s", err))
				continue L
			}
		} else {
			r.errlog(fmt.Sprintf("%s", err))
			r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds][NG] ", maxDigits, maxLength), n, id)
			if err := r.Perform(ctx, tq.task.Ng, tq.target, tq.task, q); err != nil {
				r.errlog(fmt.Sprintf("%s", err))
				continue L
			}
		}
	}

	r.logPrefix = ""
	r.log("Session finished")
	return nil
}

func (r *Runner) Perform(ctx context.Context, o *task.Operation, i *gh.Target, t *task.Task, q chan TaskQueue) error {
	if o == nil {
		return nil
	}

	r.mu.Lock()
	defer func() {
		r.revertEnv()
		r.mu.Unlock()
	}()

	os.Setenv("GHDAG_TARGET_NUMBER", fmt.Sprintf("%d", i.Number()))
	os.Setenv("GHDAG_TASK_ID", t.Id)

	for _, e := range t.Env {
		os.Setenv(e.Name, e.Value)
	}

	switch {
	case o.Run != "":
		c := exec.CommandContext(ctx, "sh", "-c", "o.Run")
		c.Env = os.Environ()
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		return nil
	case len(o.Labels) > 0:
		r.log(fmt.Sprintf("Set labels: %s", strings.Join(o.Labels, ", ")))
		return r.github.SetLabels(ctx, i.Number(), o.Labels)
	case len(o.Assignees) > 0:
		r.log(fmt.Sprintf("Set assignees: %s", strings.Join(o.Assignees, ", ")))
		return r.github.SetAssignees(ctx, i.Number(), o.Assignees)
	case o.Comment != "":
		r.log(fmt.Sprintf("Add comment: %s", o.Comment))
		return r.github.AddComment(ctx, i.Number(), o.Comment)
	case o.Action != "":
		r.log(fmt.Sprintf("%s: #%d", o.Action, i.Number()))
		switch o.Action {
		case "close":
			return r.github.CloseIssue(ctx, i.Number())
		case "merge":
			return r.github.MergePullRequest(ctx, i.Number())
		default:
			return fmt.Errorf("invalid action: %s", o.Action)
		}
	case o.Notify != "":
		r.log(fmt.Sprintf("Send notification: %s", o.Notify))
		return r.slack.PostMessage(ctx, o.Notify)
	case len(o.Next) > 0:
		r.log(fmt.Sprintf("Call next task: %s", strings.Join(o.Next, ", ")))
		for _, id := range o.Next {
			t, err := r.config.Tasks.Find(id)
			if err != nil {
				return err
			}
			q <- TaskQueue{
				target: i,
				task:   t,
				force:  true,
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

func (r *Runner) revertEnv() {
	for _, e := range os.Environ() {
		splitted := strings.Split(e, "=")
		os.Unsetenv(splitted[0])
	}
	for _, e := range r.env {
		splitted := strings.Split(e, "=")
		os.Setenv(splitted[0], splitted[1])
	}
}
