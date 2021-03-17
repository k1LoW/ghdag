package name

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCheckSyntax(t *testing.T) {
	tests := []struct {
		names LinkedNames
		want  bool
	}{
		{
			LinkedNames{},
			true,
		},
		{
			LinkedNames{
				&LinkedName{
					Github: "bob",
					Slack:  "bob_marly",
				},
			},
			true,
		},
		{
			LinkedNames{
				&LinkedName{
					Github: "bob",
					Slack:  "bob",
				},
			},
			false,
		},
		{
			LinkedNames{
				&LinkedName{
					Github: "bob",
					Slack:  "bob_marly",
				},
				&LinkedName{
					Github: "bob",
					Slack:  "bob_dylan",
				},
			},
			false,
		},
		{
			LinkedNames{
				&LinkedName{
					Github: "bob",
					Slack:  "bob_marly",
				},
				&LinkedName{
					Github: "bob_marly",
					Slack:  "bob_dylan",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		got, _ := tt.names.CheckSyntax()
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestLinkedNames(t *testing.T) {
	tests := []struct {
		names      LinkedNames
		in         []string
		wantGithub []string
		wantSlack  []string
	}{
		{
			LinkedNames{},
			[]string{"alice", "bob", "charlie"},
			[]string{"alice", "bob", "charlie"},
			[]string{"alice", "bob", "charlie"},
		},
		{
			LinkedNames{
				&LinkedName{
					Github: "bob",
					Slack:  "bob_marly",
				},
			},
			[]string{"alice", "bob", "charlie"},
			[]string{"alice", "bob", "charlie"},
			[]string{"alice", "bob_marly", "charlie"},
		},
		{
			LinkedNames{
				&LinkedName{
					Github: "bob",
					Slack:  "bob_marly",
				},
			},
			[]string{"alice", "bob_marly", "charlie"},
			[]string{"alice", "bob", "charlie"},
			[]string{"alice", "bob_marly", "charlie"},
		},
		{
			LinkedNames{
				&LinkedName{
					Github: "bob",
					Slack:  "bob_marly",
				},
			},
			[]string{"alice_liddel", "bob_marly", "charlie_sheen"},
			[]string{"alice_liddel", "bob", "charlie_sheen"},
			[]string{"alice_liddel", "bob_marly", "charlie_sheen"},
		},
	}
	for _, tt := range tests {
		gotGithub := tt.names.ToGithubNames(tt.in)
		if diff := cmp.Diff(gotGithub, tt.wantGithub, nil); diff != "" {
			t.Errorf("%s", diff)
		}
		gotSlack := tt.names.ToSlackNames(tt.in)
		if diff := cmp.Diff(gotSlack, tt.wantSlack, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
