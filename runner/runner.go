package runner

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/antonmedv/expr"
	"github.com/k1LoW/exec"
	"github.com/k1LoW/ghdag/config"
	"github.com/k1LoW/ghdag/gh"
	"github.com/k1LoW/ghdag/slk"
	"github.com/k1LoW/ghdag/task"
)

type Runner struct {
	config *config.Config
	github *gh.Client
	slack  *slk.Client
	env    []string
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
		config: c,
		github: gc,
		slack:  sc,
		env:    os.Environ(),
	}, nil
}

type TaskQueue struct {
	target *gh.Target
	task   *task.Task
	force  bool
}

func (r *Runner) Run(ctx context.Context) error {
	log.Println("session start")
	log.Printf("fetch open issues and pull requests from %s", os.Getenv("GITHUB_REPOSITORY"))
	targets, err := r.github.FetchTargets(ctx)
	if err != nil {
		return err
	}
	log.Printf("%d issues/pull requests are fetched\n", len(targets))
	tasks := r.config.Tasks
	log.Printf("%d tasks are loaded\n", len(tasks))

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

		prefix := fmt.Sprintf("[#%d <- %s] ", tq.target.Number(), tq.task.Id)

		if tq.task.If != "" {
			do, err := expr.Eval(fmt.Sprintf("(%s) == true", tq.task.If), tq.target.Dump())
			if err != nil {
				log.Printf("%s%s\n", prefix, err)
				continue L
			}
			if !do.(bool) {
				continue L
			}
		}
		if tq.task.If == "" && !tq.force {
			continue L
		}

		log.Printf("%s%s\n", prefix, "perform `do:` operation")
		if err := r.Perform(ctx, tq.task.Do, tq.target, tq.task, q); err == nil {
			log.Printf("%s%s\n", prefix, "perform `ok:` operation")
			if err := r.Perform(ctx, tq.task.Ok, tq.target, tq.task, q); err != nil {
				log.Printf("%serror: %s\n", prefix, err)
				continue L
			}
		} else {
			log.Printf("%serror: %s\n", prefix, err)
			log.Printf("%s%s\n", prefix, "perform `ng:` operation")
			if err := r.Perform(ctx, tq.task.Ng, tq.target, tq.task, q); err != nil {
				log.Printf("%serror: %s\n", prefix, err)
				continue L
			}
		}
	}

	log.Println("session finished")
	return nil
}

func (r *Runner) Perform(ctx context.Context, o *task.Operation, i *gh.Target, t *task.Task, q chan TaskQueue) error {
	if o == nil {
		return nil
	}

	prefix := fmt.Sprintf("[#%d <- %s] ", i.Number(), t.Id)

	env := r.env
	for _, e := range t.Env {
		env = append(env, fmt.Sprintf("%s=%s", e.Name, e.Value))
	}

	switch {
	case o.Run != "":
		c := exec.CommandContext(ctx, "sh", "-c", "o.Run")
		c.Env = env
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		return nil
	case len(o.Labels) > 0:
		log.Printf("%sset labels: %s", prefix, o.Labels)
		return r.github.SetLabels(ctx, i.Number(), o.Labels)
	case len(o.Assignees) > 0:
		log.Printf("%sset assignees: %s", prefix, o.Assignees)
		return r.github.SetAssignees(ctx, i.Number(), o.Labels)
	case o.Comment != "":
		log.Printf("%sadd comment: %s", prefix, o.Comment)
		return r.github.AddComment(ctx, i.Number(), o.Comment)
	case o.Action != "":
		log.Printf("%s%s: #%d", prefix, o.Action, i.Number())
		switch o.Action {
		case "close":
			return r.github.CloseIssue(ctx, i.Number())
		case "merge":
			return r.github.MergePullRequest(ctx, i.Number())
		default:
			return fmt.Errorf("invalid action: %s", o.Action)
		}
	case o.Notify != "":
		return fmt.Errorf("not implemented: %s", "notify")
	case len(o.Next) > 0:
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
