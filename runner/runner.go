package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antonmedv/expr"
	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/exec"
	"github.com/k1LoW/ghdag/config"
	"github.com/k1LoW/ghdag/env"
	"github.com/k1LoW/ghdag/erro"
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
		r.logPrefix = ""
		r.log("Session finished")
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

	targets, err := r.fetchTargets(ctx)
	maxDigits := targets.MaxDigits()
	r.log(fmt.Sprintf("%d issues and pull requests are fetched", len(targets)))
	if errors.As(err, &erro.NotOpenError{}) {
		r.log(fmt.Sprintf("[SKIP] %s", err))
		return nil
	}
	if err != nil {
		return err
	}
	tasks := r.config.Tasks
	r.log(fmt.Sprintf("%d tasks are loaded", len(tasks)))
	maxLength := tasks.MaxLengthID()

	q := make(chan TaskQueue, len(tasks)*len(targets)+100)
	for _, i := range targets {
		for _, t := range tasks {
			q <- TaskQueue{
				target: i,
				task:   t,
			}
		}
	}

	for {
		if len(q) == 0 {
			close(q)
		}

		tq, ok := <-q
		if !ok {
			break
		}

		err := func() error {
			r.mu.Lock()
			defer func() {
				_ = r.revertEnv()
				r.mu.Unlock()
			}()

			n := tq.target.Number
			id := tq.task.Id
			r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] ", maxDigits, maxLength), n, id)

			dump := tq.target.Dump()
			for k, v := range dump {
				ek := strings.ToUpper(fmt.Sprintf("GHDAG_TARGET_%s", k))
				switch v := v.(type) {
				case bool:
					ev := "true"
					if !v {
						ev = "false"
					}
					if err := os.Setenv(ek, ev); err != nil {
						return err
					}
				case float64:
					if err := os.Setenv(ek, fmt.Sprintf("%g", v)); err != nil {
						return err
					}
				case string:
					if err := os.Setenv(ek, v); err != nil {
						return err
					}
				case []interface{}:
					ev := []string{}
					for _, i := range v {
						ev = append(ev, i.(string))
					}
					if err := os.Setenv(ek, strings.Join(ev, ", ")); err != nil {
						return err
					}
				}
			}
			if err := os.Setenv("GHDAG_TASK_ID", id); err != nil {
				return err
			}
			if err := r.config.Env.Setenv(); err != nil {
				return err
			}
			if err := tq.task.Env.Setenv(); err != nil {
				return err
			}

			if tq.task.If != "" {
				do, err := expr.Eval(fmt.Sprintf("(%s) == true", tq.task.If), dump)
				if err != nil {
					r.errlog(fmt.Sprintf("%s", err))
					return nil
				}
				if !do.(bool) {
					r.debuglog(fmt.Sprintf("[SKIP] the condition in the `if` section is not met (%s)", tq.task.If))
					return nil
				}
			}
			if tq.task.If == "" && !tq.called {
				r.debuglog("[SKIP] the `if:` section is missing")
				return nil
			}

			if tq.called {
				// Update target
				target, err := r.github.FetchTarget(ctx, tq.target.Number)
				if err != nil {
					if errors.As(err, &erro.NotOpenError{}) {
						r.log(fmt.Sprintf("[SKIP] %s", err))
						return nil
					}
					return err
				}
				tq.target = target
			}

			r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [DO] ", maxDigits, maxLength), n, id)
			if err := r.perform(ctx, tq.task.Do, tq.target, tq.task, q); err == nil {
				r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [OK] ", maxDigits, maxLength), n, id)
				if err := r.perform(ctx, tq.task.Ok, tq.target, tq.task, q); err != nil {
					if errors.As(err, &erro.AlreadyInStateError{}) {
						r.log(fmt.Sprintf("[SKIP] %s", err))
						return nil
					}
					r.errlog(fmt.Sprintf("%s", err))
					return nil
				}
			} else {
				if errors.As(err, &erro.AlreadyInStateError{}) {
					r.log(fmt.Sprintf("[SKIP] %s", err))
					return nil
				}
				r.errlog(fmt.Sprintf("%s", err))
				if err := os.Setenv("GHDAG_ACTION_OK_ERROR", fmt.Sprintf("%s", err)); err != nil {
					return err
				}
				r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [NG] ", maxDigits, maxLength), n, id)
				if err := r.perform(ctx, tq.task.Ng, tq.target, tq.task, q); err != nil {
					if errors.As(err, &erro.AlreadyInStateError{}) {
						r.log(fmt.Sprintf("[SKIP] %s", err))
						return nil
					}
					r.errlog(fmt.Sprintf("%s", err))
					return nil
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) perform(ctx context.Context, a *task.Action, i *target.Target, t *task.Task, q chan TaskQueue) error {
	if a == nil {
		return nil
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
		sortStringSlice(i.Labels)
		sortStringSlice(a.Labels)
		r.log(fmt.Sprintf("Set labels: %s", strings.Join(a.Labels, ", ")))
		if cmp.Equal(i.Labels, a.Labels) {
			return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(a.Labels, ", ")))
		}
		return r.github.SetLabels(ctx, i.Number, a.Labels)
	case len(a.Assignees) > 0:
		as, err := r.github.ResolveUsers(ctx, a.Assignees)
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
			return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(a.Assignees, ", ")))
		}
		return r.github.SetAssignees(ctx, i.Number, as)
	case len(a.Reviewers) > 0:
		reviewers, err := r.sampleByEnv(a.Reviewers, "GITHUB_REVIEWERS_SAMPLE")
		if err != nil {
			return err
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
			return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", strings.Join(a.Reviewers, ", ")))
		}
		return r.github.SetReviewers(ctx, i.Number, ra)
	case a.Comment != "":
		c, err := env.ParseWithEnviron(a.Comment, env.EnvMap())
		if err != nil {
			return err
		}
		r.log(fmt.Sprintf("Add comment: %s", c))

		if i.NumberOfConsecutiveComments >= 5 {
			return fmt.Errorf("Too many comments in a row by ghdag: %d", i.NumberOfConsecutiveComments)
		}

		c = fmt.Sprintf("%s\n%s%s:%s -->\n", c, gh.CommentLogPrefix, t.Id, a.Type)

		if i.LatestCommentBody == c {
			return erro.NewAlreadyInStateError(fmt.Errorf("the target is already in a state of being wanted: %s", c))
		}
		return r.github.AddComment(ctx, i.Number, c)
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
		n, err := env.ParseWithEnviron(a.Notify, env.EnvMap())
		if err != nil {
			return err
		}
		mentions := strings.Split(os.Getenv("SLACK_MENTIONS"), " ")
		mentions, err = r.sampleByEnv(mentions, "SLACK_MENTIONS_SAMPLE")
		if err != nil {
			return err
		}
		r.log(fmt.Sprintf("Send notification: %s", n))
		return r.slack.PostMessage(ctx, n, mentions)
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

func (r *Runner) fetchTargets(ctx context.Context) (target.Targets, error) {
	en := os.Getenv("GITHUB_EVENT_NAME")
	ep := os.Getenv("GITHUB_EVENT_PATH")
	if strings.HasPrefix(en, "issue") || strings.HasPrefix(en, "pull_request") {
		n, state, err := detectTargetNumber(ep)
		if err != nil {
			return nil, err
		}
		if state != "open" {
			return nil, erro.NewNotOpenError(fmt.Errorf("#%d is not opened: %s", n, state))
		}
		r.log(fmt.Sprintf("Fetch #%d from %s", n, os.Getenv("GITHUB_REPOSITORY")))
		t, err := r.github.FetchTarget(ctx, n)
		if err != nil {
			return nil, err
		}
		return target.Targets{n: t}, nil
	}
	r.log(fmt.Sprintf("Fetch all open issues and pull requests from %s", os.Getenv("GITHUB_REPOSITORY")))
	return r.github.FetchTargets(ctx)
}

func (r *Runner) sampleByEnv(in []string, envKey string) ([]string, error) {
	if os.Getenv(envKey) == "" {
		return in, nil
	}
	r.debuglog(fmt.Sprintf("found environment variable for sampling: %s", envKey))
	sn, err := strconv.Atoi(os.Getenv(envKey))
	if err != nil {
		return nil, err
	}
	if len(in) > sn {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(in), func(i, j int) { in[i], in[j] = in[j], in[i] })
		in = in[:sn]
	}
	return in, nil
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

func detectTargetNumber(p string) (int, string, error) {
	b, err := ioutil.ReadFile(filepath.Clean(p))
	if err != nil {
		return 0, "", err
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
		return 0, "", err
	}
	switch {
	case s.PullRequest.Number > 0:
		return s.PullRequest.Number, s.PullRequest.State, nil
	case s.Issue.Number > 0:
		return s.Issue.Number, s.Issue.State, nil
	}
	return 0, "", fmt.Errorf("can not parse: %s", p)
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func sortStringSlice(in []string) {
	sort.Slice(in, func(i, j int) bool {
		return in[i] < in[j]
	})
}
