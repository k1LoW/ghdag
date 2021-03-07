package runner

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/k1LoW/ghdag/erro"
	"github.com/k1LoW/ghdag/mock"
	"github.com/k1LoW/ghdag/target"
)

func TestPerformRunAction(t *testing.T) {
	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.revertEnv(); err != nil {
			t.Fatal(err)
		}
	}()

	tests := []struct {
		in         string
		wantStdout string
		wantStderr string
		wantErr    bool
	}{
		{"echo hello", "hello\n", "", false},
		{"echo world 1>&2", "", "world\n", false},
		{"unknowncmd", "", "", true},
	}
	for _, tt := range tests {
		if err := r.revertEnv(); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		i := &target.Target{}
		if err := faker.FakeData(i); err != nil {
			t.Fatal(err)
		}
		if err := r.PerformRunAction(ctx, i, tt.in); (err == nil) == tt.wantErr {
			t.Errorf("got %v\nwantErr %v", err, tt.wantErr)
		}
		if got := os.Getenv("GHDAG_ACTION_RUN_STDOUT"); got != tt.wantStdout {
			t.Errorf("got %v\nwant %v", got, tt.wantStdout)
		}
		if got := os.Getenv("GHDAG_ACTION_RUN_STDERR"); got != tt.wantStderr {
			t.Errorf("got %v\nwant %v", got, tt.wantStdout)
		}
	}
}

func TestPerformLabelsAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.revertEnv(); err != nil {
			t.Fatal(err)
		}
	}()

	m := mock.NewMockGitHubClient(ctrl)
	r.github = m

	tests := []struct {
		in      []string
		current []string
		want    []string
		wantErr interface{}
	}{
		{[]string{"bug", "question"}, nil, []string{"bug", "question"}, nil},
		{[]string{"bug", "question"}, []string{"bug", "question"}, []string{"bug", "question"}, &erro.AlreadyInStateError{}},
	}
	for _, tt := range tests {
		if err := r.revertEnv(); err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		i := &target.Target{}
		if err := faker.FakeData(i); err != nil {
			t.Fatal(err)
		}
		if tt.current != nil {
			i.Labels = tt.current
		}
		if tt.wantErr == nil {
			m.EXPECT().SetLabels(gomock.Eq(ctx), gomock.Eq(i.Number), gomock.Eq(tt.want)).Return(nil)
			if err := r.PerformLabelsAction(ctx, i, tt.in); err != nil {
				t.Error(err)
			}
		} else {
			if err := r.PerformLabelsAction(ctx, i, tt.in); !errors.As(err, tt.wantErr) {
				t.Errorf("got %v\nwant %v", err, tt.wantErr)
			}
		}
	}
}
