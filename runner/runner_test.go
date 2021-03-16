package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/ghdag/env"
	"github.com/k1LoW/ghdag/target"
)

func TestCheckIf(t *testing.T) {
	envCache := os.Environ()
	defer func() {
		if err := env.Revert(envCache); err != nil {
			t.Fatal(err)
		}
	}()
	tests := []struct {
		cond string
		env  map[string]string
		want bool
	}{
		{
			"",
			map[string]string{},
			false,
		},
		{
			"github.event_name == 'issues'",
			map[string]string{
				"GITHUB_EVENT_NAME": "issues",
			},
			true,
		},
		{
			"'bug' in caller_action_labels_updated",
			map[string]string{
				"GHDAG_ACTION_LABELS_UPDATED": "bug question",
			},
			true,
		},
		{
			`github.event_name
		==
		'issues'
		&&
		'question'
		in
		caller_action_labels_updated`,
			map[string]string{
				"GITHUB_EVENT_NAME":           "issues",
				"GHDAG_ACTION_LABELS_UPDATED": "bug question",
			},
			true,
		},
		{
			`github.event_name == 'issues'
		&& github.event.action == 'opened'
		&& github.event.issue.state == 'open'`,
			map[string]string{
				"GITHUB_EVENT_NAME": "issues",
				"GITHUB_EVENT_PATH": filepath.Join(testdataDir(), "event_issue_opened.json"),
			},
			true,
		},
	}
	for _, tt := range tests {
		if err := env.Revert(envCache); err != nil {
			t.Fatal(err)
		}
		for k, v := range tt.env {
			os.Setenv(k, v)
		}
		r, err := New(nil)
		if err != nil {
			t.Fatal(err)
		}
		i := &target.Target{}
		if err := faker.FakeData(i); err != nil {
			t.Fatal(err)
		}
		got := r.CheckIf(tt.cond, i)
		if got != tt.want {
			t.Errorf("if(%s) got %v\nwant %v", tt.cond, got, tt.want)
		}
	}
}

func TestDetectTargetNumber(t *testing.T) {
	tests := []struct {
		path       string
		wantNumber int
		wantState  string
		wantErr    bool
	}{
		{"event_issue_opened.json", 19, "open", false},
		{"event_pull_request_opened.json", 20, "open", false},
		{"event_issue_comment_opened.json", 20, "open", false},
	}
	envCache := os.Environ()
	for _, tt := range tests {
		if err := env.Revert(envCache); err != nil {
			t.Fatal(err)
		}
		os.Setenv("GITHUB_EVENT_PATH", filepath.Join(testdataDir(), tt.path))
		got, err := decodeGitHubEvent()
		if tt.wantErr && err != nil {
			continue
		}
		if err != nil {
			t.Error(err)
		}
		if got.Number != tt.wantNumber {
			t.Errorf("got %v\nwant %v", got.Number, tt.wantNumber)
		}
		if got.State != tt.wantState {
			t.Errorf("got %v\nwant %v", got.State, tt.wantState)
		}
	}
}

func TestSampleByEnv(t *testing.T) {
	tests := []struct {
		env  int
		want int
	}{
		{3, 3},
		{2, 2},
		{4, 3},
		{0, 0},
	}
	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		r.initSeed()
		in := []string{"alice", "bob", "charlie"}
		envKey := "TEST_SAMPLE_BY_ENV"
		os.Setenv(envKey, fmt.Sprintf("%d", tt.env))
		got, err := r.sample(in, envKey)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestSampleByEnvWithSameSeed(t *testing.T) {
	tests := []struct {
		enable bool
		diff   bool
	}{
		{false, true},
		{true, false},
	}
	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		if err := r.revertEnv(); err != nil {
			t.Fatal(err)
		}
		if err := os.Setenv("GHDAG_SAMPLE_WITH_SAME_SEED", fmt.Sprintf("%t", tt.enable)); err != nil {
			t.Fatal(err)
		}
		a := []string{}
		b := []string{}
		for i := 0; i < 100; i++ {
			a = append(a, fmt.Sprintf("%d", i))
			b = append(b, fmt.Sprintf("%d", i))
		}
		envKey := "TEST_SAMPLE_BY_ENV"
		os.Setenv(envKey, "99")

		r.initSeed()
		got, err := r.sample(a, envKey)
		if err != nil {
			t.Fatal(err)
		}

		r.initSeed()
		got2, err := r.sample(b, envKey)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(got, got2, nil); (diff != "") != tt.diff {
			t.Error("sample error")
		}
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
