package runner

import (
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

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
