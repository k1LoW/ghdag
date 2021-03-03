package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
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
		in := []string{"alice", "bob", "charlie"}
		envKey := "TEST_SAMPLE_BY_ENV"
		os.Setenv(envKey, fmt.Sprintf("%d", tt.env))
		got, err := r.sampleByEnv(in, envKey)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
