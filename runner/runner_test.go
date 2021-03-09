package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDetectTargetNumber(t *testing.T) {
	tests := []struct {
		path      string
		want      int
		wantState string
		wantErr   bool
	}{
		{"event_issue_opened.json", 19, "open", false},
		{"event_pull_request_opened.json", 20, "open", false},
		{"event_issue_comment_opened.json", 20, "open", false},
	}
	for _, tt := range tests {
		got, gotState, err := detectTargetNumber(filepath.Join(testdataDir(), tt.path))
		if tt.wantErr && err != nil {
			continue
		}
		if err != nil {
			t.Error(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
		if gotState != tt.wantState {
			t.Errorf("got %v\nwant %v", gotState, tt.wantState)
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
