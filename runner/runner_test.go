package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTargetNumber(t *testing.T) {
	tests := []struct {
		path    string
		want    int
		wantErr bool
	}{
		{"event_issue_opened.json", 19, false},
		{"event_pull_request_opened.json", 20, false},
		{"event_issue_comment_opened.json", 20, false},
	}
	for _, tt := range tests {
		got, err := detectTargetNumber(filepath.Join(testdataDir(), tt.path))
		if tt.wantErr && err != nil {
			continue
		}
		if err != nil {
			t.Error(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func testdataDir() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	return dir
}
