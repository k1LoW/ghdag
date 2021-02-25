package env

import (
	"os"
	"strings"
	"testing"
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
				"GHDAG_TOKEN": "zzz${ TOKEN }zzz",
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

		after := envMap()

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

func clearEnv() {
	for _, e := range os.Environ() {
		splitted := strings.Split(e, "=")
		os.Unsetenv(splitted[0])
	}
}
