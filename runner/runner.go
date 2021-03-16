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
	mu         sync.Mutex
	config     *config.Config
	github     gh.GhClient
	slack      slk.SlkClient
	event      *GitHubEvent
	envCache   []string
	logPrefix  string
	seed       int64
	excludeKey int
}

func New(c *config.Config) (*Runner, error) {
	e, _ := decodeGitHubEvent()
	return &Runner{
		config:     c,
		github:     nil,
		slack:      nil,
		event:      e,
		envCache:   os.Environ(),
		logPrefix:  "",
		seed:       time.Now().UnixNano(),
		excludeKey: -1,
	}, nil
}

type TaskQueue struct {
	target           *target.Target
	task             *task.Task
	called           bool
	callerTask       *task.Task
	callerSeed       int64
	callerExcludeKey int
	callerEnv        env.Env
}

func (r *Runner) Run(ctx context.Context) error {
	r.logPrefix = ""
	r.log("Start session")
	r.log(fmt.Sprintf("github.event_name: %s", r.event.Name))
	r.debuglog(fmt.Sprintf("github.event: %s", r.event.RawPayload))
	defer func() {
		_ = r.revertEnv()
		r.logPrefix = ""
		r.log("Session finished")
	}()
	if err := r.config.Env.Setenv(); err != nil {
		return err
	}

	if err := r.InitClients(); err != nil {
		return err
	}

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

			if err := r.initTaskEnv(tq); err != nil {
				return err
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

				// Set task id of caller
				if err := os.Setenv("GHDAG_CALLER_TASK_ID", tq.callerTask.Id); err != nil {
					return err
				}

				// Set caller seed
				r.seed = tq.callerSeed
				r.excludeKey = tq.callerExcludeKey
			}

			if tq.task.If != "" {
				if !r.CheckIf(tq.task.If, tq.target) {
					return nil
				}
			} else {
				if !tq.called {
					r.debuglog("[SKIP] the `if:` section is missing")
					return nil
				}
			}

			r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [DO] ", maxDigits, maxLength), n, id)
			if err := r.perform(ctx, tq.task.Do, tq.target, tq.task, q); err == nil {
				r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [OK] ", maxDigits, maxLength), n, id)
				if err := r.perform(ctx, tq.task.Ok, tq.target, tq.task, q); err != nil {
					r.initSeed()
					if errors.As(err, &erro.AlreadyInStateError{}) || errors.As(err, &erro.NoReviewerError{}) {
						r.log(fmt.Sprintf("[SKIP] %s", err))
						return nil
					}
					r.errlog(fmt.Sprintf("%s", err))
					return nil
				}
			} else {
				if errors.As(err, &erro.AlreadyInStateError{}) || errors.As(err, &erro.NoReviewerError{}) {
					r.log(fmt.Sprintf("[SKIP] %s", err))
					return nil
				}
				r.errlog(fmt.Sprintf("%s", err))
				if err := os.Setenv("GHDAG_ACTION_DO_ERROR", fmt.Sprintf("%s", err)); err != nil {
					return err
				}
				r.logPrefix = fmt.Sprintf(fmt.Sprintf("[#%%-%dd << %%-%ds] [NG] ", maxDigits, maxLength), n, id)
				if err := r.perform(ctx, tq.task.Ng, tq.target, tq.task, q); err != nil {
					if errors.As(err, &erro.AlreadyInStateError{}) || errors.As(err, &erro.NoReviewerError{}) {
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

func (r *Runner) InitClients() error {
	if r.github == nil {
		gc, err := gh.NewClient()
		if err != nil {
			return err
		}
		r.github = gc
	}
	if r.slack == nil {
		sc, err := slk.NewClient()
		if err != nil {
			return err
		}
		r.slack = sc
	}
	return nil
}

func (r *Runner) CheckIf(cond string, i *target.Target) bool {
	if cond == "" {
		return false
	}
	isCalled := true
	k := "GHDAG_TASK_IS_CALLED"
	if os.Getenv(k) == "" || strings.ToLower(os.Getenv(k)) == "false" || os.Getenv(k) == "0" {
		isCalled = false
	}
	now := time.Now()
	variables := map[string]interface{}{
		"year":      now.UTC().Year(),
		"month":     now.UTC().Month(),
		"day":       now.UTC().Day(),
		"hour":      now.UTC().Hour(),
		"weekday":   int(now.UTC().Weekday()),
		"is_called": isCalled,
		"github": map[string]interface{}{
			"event_name": r.event.Name,
			"event":      r.event.Payload,
		},
	}
	for _, k := range propagatableEnv {
		v := os.Getenv(k)
		key := strings.ToLower(strings.Replace(k, "GHDAG_", "CALLER_", 1))
		switch k {
		case "GHDAG_ACTION_LABELS_UPDATED", "GHDAG_ACTION_ASSIGNEES_UPDATED", "GHDAG_ACTION_REVIEWERS_UPDATED":
			a, _ := env.Split(v)
			variables[key] = a
		default:
			variables[key] = v
		}
	}
	variables = merge(variables, i.Dump())
	doOrNot, err := expr.Eval(fmt.Sprintf("(%s) == true", cond), variables)
	if err != nil {
		r.errlog(fmt.Sprintf("%s", err))
		return false
	}
	if !doOrNot.(bool) {
		r.debuglog(fmt.Sprintf("[SKIP] the condition in the `if` section is not met (%s)", cond))
		return false
	}
	return true
}

func (r *Runner) perform(ctx context.Context, a *task.Action, i *target.Target, t *task.Task, q chan TaskQueue) error {
	if a == nil {
		return nil
	}
	r.initSeed()

	switch {
	case a.Run != "":
		return r.PerformRunAction(ctx, i, a.Run)
	case len(a.Labels) > 0:
		return r.PerformLabelsAction(ctx, i, a.Labels)
	case len(a.Assignees) > 0 || (a.Assignees != nil && os.Getenv("GITHUB_ASSIGNEES") != ""):
		as, err := env.Split(os.Getenv("GITHUB_ASSIGNEES"))
		if err != nil {
			return err
		}
		assignees := append(a.Assignees, as...)
		return r.PerformAssigneesAction(ctx, i, assignees)
	case len(a.Reviewers) > 0 || (a.Reviewers != nil && os.Getenv("GITHUB_REVIEWERS") != ""):
		rs, err := env.Split(os.Getenv("GITHUB_ASSIGNEES"))
		if err != nil {
			return err
		}
		reviewers := append(a.Reviewers, rs...)
		return r.PerformReviewersAction(ctx, i, reviewers)
	case a.Comment != "":
		return r.PerformCommentAction(ctx, i, a.Comment)
	case a.State != "":
		return r.PerformStateAction(ctx, i, a.State)
	case a.Notify != "":
		return r.PerformNotifyAction(ctx, i, a.Notify)
	case len(a.Next) > 0:
		return r.performNextAction(ctx, i, t, q, a.Next)
	}
	return nil
}

func (r *Runner) initSeed() {
	k := "GHDAG_SAMPLE_WITH_SAME_SEED"
	if os.Getenv(k) == "" || strings.ToLower(os.Getenv(k)) == "false" || os.Getenv(k) == "0" {
		r.seed = time.Now().UnixNano()
		r.excludeKey = -1
	}
}

func (r *Runner) initTaskEnv(tq TaskQueue) error {
	id := tq.task.Id
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

	var isCalled string
	if tq.called {
		isCalled = "1"
	} else {
		isCalled = "0"
	}
	if err := os.Setenv("GHDAG_TASK_IS_CALLED", isCalled); err != nil {
		return err
	}
	if err := r.config.Env.Setenv(); err != nil {
		return err
	}
	if err := tq.task.Env.Setenv(); err != nil {
		return err
	}
	if tq.called {
		if err := tq.callerEnv.Setenv(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) fetchTargets(ctx context.Context) (target.Targets, error) {
	en := os.Getenv("GITHUB_EVENT_NAME")
	if strings.HasPrefix(en, "issue") || strings.HasPrefix(en, "pull_request") {
		t, err := r.FetchTarget(ctx, 0)
		if err != nil {
			return nil, err
		}
		return target.Targets{t.Number: t}, nil
	}
	r.log(fmt.Sprintf("Fetch all open issues and pull requests from %s", os.Getenv("GITHUB_REPOSITORY")))
	return r.github.FetchTargets(ctx)
}

func (r *Runner) FetchTarget(ctx context.Context, n int) (*target.Target, error) {
	if n > 0 {
		return r.github.FetchTarget(ctx, n)
	}
	if !strings.HasPrefix(r.event.Name, "issue") && !strings.HasPrefix(r.event.Name, "pull_request") {
		return nil, fmt.Errorf("unsupported event: %s", r.event.Name)
	}
	if r.event.State != "open" {
		return nil, erro.NewNotOpenError(fmt.Errorf("#%d is %s", n, r.event.State))
	}
	r.log(fmt.Sprintf("Fetch #%d from %s", r.event.Number, os.Getenv("GITHUB_REPOSITORY")))
	return r.github.FetchTarget(ctx, r.event.Number)
}

func (r *Runner) setExcludeKey(in []string, exclude string) error {
	for k, v := range in {
		if v == exclude {
			r.excludeKey = k
			return nil
		}
	}
	return fmt.Errorf("not found key: %s", exclude)
}

func (r *Runner) sample(in []string, envKey string) ([]string, error) {
	if r.excludeKey >= 0 {
		in = unset(in, r.excludeKey)
	}
	if os.Getenv(envKey) == "" {
		return in, nil
	}
	r.debuglog(fmt.Sprintf("env %s is set for sampling", envKey))
	sn, err := strconv.Atoi(os.Getenv(envKey))
	if err != nil {
		return nil, err
	}

	if len(in) > sn {
		rand.Seed(r.seed)
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

type GitHubEvent struct {
	Name       string
	Number     int
	State      string
	Payload    interface{}
	RawPayload []byte
}

func decodeGitHubEvent() (*GitHubEvent, error) {
	p := os.Getenv("GITHUB_EVENT_PATH")
	n := os.Getenv("GITHUB_EVENT_NAME")
	i := &GitHubEvent{
		Name: n,
	}
	b, err := ioutil.ReadFile(filepath.Clean(p))
	if err != nil {
		return i, err
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
		return i, err
	}
	switch {
	case s.PullRequest.Number > 0:
		i.Number = s.PullRequest.Number
		i.State = s.PullRequest.State
	case s.Issue.Number > 0:
		i.Number = s.Issue.Number
		i.State = s.Issue.State
	}

	var payload interface{}

	if err := json.Unmarshal(b, &payload); err != nil {
		return i, err
	}

	i.Payload = payload
	i.RawPayload = b

	return i, nil
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

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func unset(s []string, i int) []string {
	if i >= len(s) {
		return s
	}
	return append(s[:i], s[i+1:]...)
}

func exclude(s []string, e string) []string {
	o := []string{}
	for _, v := range s {
		if v == e {
			continue
		}
		o = append(o, v)
	}
	return o
}

func merge(ms ...map[string]interface{}) map[string]interface{} {
	o := map[string]interface{}{}
	for _, m := range ms {
		for k, v := range m {
			o[k] = v
		}
	}
	return o
}

func sortStringSlice(in []string) {
	sort.Slice(in, func(i, j int) bool {
		return in[i] < in[j]
	})
}
