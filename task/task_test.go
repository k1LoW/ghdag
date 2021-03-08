package task

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/ghdag/env"
)

func TestCheckSyntax(t *testing.T) {
	tests := []struct {
		in     []byte
		env    map[string]string
		wantOk bool
	}{
		{[]byte(`
id: task-id
if: bug in labels
do:
  assignees: [alice bob charlie]
`), map[string]string{}, true},
		{[]byte(`
id: task-id
if: bug in labels
do:
`), map[string]string{}, false},
		{[]byte(`
id: task-id
if: bug in labels
do:
  assignees: [alice bob charlie]
  comment: hello
`), map[string]string{}, false},
		{[]byte(`
id: task-id
if: bug in labels
do:
env:
  GITHUB_ASSIGNEES: alice bob charlie
`), map[string]string{}, false},
		{[]byte(`
id: task-id
if: bug in labels
do:
  assignees: []
`), map[string]string{}, false},
		{[]byte(`
id: task-id
if: bug in labels
do:
  assignees: []
env:
  GITHUB_ASSIGNEES: alice bob charlie
`), map[string]string{}, true},
		{[]byte(`
id: task-id
if: bug in labels
do:
  reviewers: []
env:
  GITHUB_REVIEWERS: alice bob charlie
`), map[string]string{}, true},
		{[]byte(`
id: task-id
if: bug in labels
do:
  assignees: []
`), map[string]string{
			"GITHUB_ASSIGNEES": "alice bob charlie",
		}, true},
		{[]byte(`
id: task-id
if: bug in labels
do:
  reviewers: []
`), map[string]string{
			"GITHUB_REVIEWERS": "alice bob charlie",
		}, true},
	}
	envCache := os.Environ()
	for _, tt := range tests {
		if err := env.Revert(envCache); err != nil {
			t.Fatal(err)
		}
		tsk := &Task{}
		if err := yaml.Unmarshal(tt.in, tsk); err != nil {
			t.Fatal(err)
		}
		for k, v := range tt.env {
			os.Setenv(k, v)
		}
		if ok, _ := tsk.CheckSyntax(); ok != tt.wantOk {
			t.Errorf("%s\ngot %v\nwant %v", tt.in, ok, tt.wantOk)
		}
	}
}
