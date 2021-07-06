package gh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/ghdag/env"
)

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
		os.Setenv("GITHUB_EVENT_NAME", "test")
		os.Setenv("GITHUB_EVENT_PATH", filepath.Join(testdataDir(), tt.path))
		got, err := DecodeGitHubEvent()
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

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
