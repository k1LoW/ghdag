package env

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	envCache := os.Environ()

	m.Run()

	_ = Revert(envCache)
}

func TestSetenv(t *testing.T) {
	tests := []struct {
		before map[string]string
		in     Env
		want   map[string]string
	}{
		{
			map[string]string{
				"TOKEN": "abcdef",
			},
			Env{
				"GHDAG_TOKEN": "${TOKEN}",
			},
			map[string]string{
				"TOKEN":       "abcdef",
				"GHDAG_TOKEN": "abcdef",
			},
		},
		{
			map[string]string{
				"TOKEN": "abcdef",
			},
			Env{
				"GHDAG_TOKEN": "zzz${TOKEN}zzz",
			},
			map[string]string{
				"TOKEN":       "abcdef",
				"GHDAG_TOKEN": "zzzabcdefzzz",
			},
		},
	}

	for _, tt := range tests {
		clearEnv()
		for k, v := range tt.before {
			os.Setenv(k, v)
		}

		tt.in.Setenv()

		after := EnvMap()

		if len(after) != len(tt.want) {
			t.Errorf("got %v\nwant %v", len(after), len(tt.want))
		}

		for k, v := range tt.want {
			got, ok := after[k]
			if !ok {
				t.Errorf("got %v\nwant %v", ok, true)
			}
			if got != v {
				t.Errorf("got %v\nwant %v", got, v)
			}
		}
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"bug", []string{"bug"}},
		{"bug question", []string{"bug", "question"}},
		{"bug  question", []string{"bug", "question"}},
		{"bug,question", []string{"bug", "question"}},
		{"bug, question", []string{"bug", "question"}},
		{"bug,  question", []string{"bug", "question"}},
		{`bug "help wanted"`, []string{"bug", "help wanted"}},
		{`bug 'help wanted'`, []string{"bug", "help wanted"}},
		{"", []string{}},
	}
	for _, tt := range tests {
		got, err := Split(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
		got2, err := Split(Join(got))
		if diff := cmp.Diff(got, got2, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func clearEnv() {
	for _, e := range os.Environ() {
		splitted := strings.Split(e, "=")
		os.Unsetenv(splitted[0])
	}
}
