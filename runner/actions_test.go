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

	m := mock.NewMockGhClient(ctrl)
	r.github = m

	tests := []struct {
		in      []string
		current []string
		want    []string
		wantErr interface{}
	}{
		{[]string{"bug", "question"}, nil, []string{"bug", "question"}, nil},
		{[]string{"bug", "question"}, []string{"bug", "question"}, nil, &erro.AlreadyInStateError{}},
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

func TestPerformAssigneesAction(t *testing.T) {
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

	m := mock.NewMockGhClient(ctrl)
	r.github = m

	tests := []struct {
		in      []string
		current []string
		want    []string
		wantErr interface{}
	}{
		{[]string{"alice", "bob"}, nil, []string{"alice", "bob"}, nil},
		{[]string{"alice", "bob"}, []string{"alice", "bob"}, nil, &erro.AlreadyInStateError{}},
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
			i.Assignees = tt.current
		}
		m.EXPECT().ResolveUsers(gomock.Eq(ctx), gomock.Eq(tt.in)).Return(tt.in, nil)
		if tt.wantErr == nil {
			m.EXPECT().SetAssignees(gomock.Eq(ctx), gomock.Eq(i.Number), gomock.Eq(tt.want)).Return(nil)
		}
		if err := r.PerformAssigneesAction(ctx, i, tt.in); err != nil {
			if !errors.As(err, tt.wantErr) {
				t.Errorf("got %v\nwant %v", err, tt.wantErr)
			}
		}
	}
}

func TestPerformReviewersAction(t *testing.T) {
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

	m := mock.NewMockGhClient(ctrl)
	r.github = m

	tests := []struct {
		in                []string
		author            string
		current           []string
		currentCodeOwners []string
		want              []string
		wantErr           interface{}
	}{
		{[]string{"alice", "bob"}, "", nil, nil, []string{"alice", "bob"}, nil},
		{[]string{"alice", "bob"}, "", []string{"alice", "bob"}, nil, nil, &erro.AlreadyInStateError{}},
		{[]string{"alice", "bob"}, "", []string{"alice"}, nil, []string{"alice", "bob"}, nil},
		{[]string{"alice", "bob"}, "", []string{}, []string{"bob"}, []string{"alice"}, nil},
		{[]string{"alice", "bob"}, "alice", nil, nil, []string{"bob"}, nil},
		{[]string{"alice"}, "alice", nil, nil, nil, &erro.NoReviewerError{}},
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
		if tt.author != "" {
			i.Author = tt.author
		}
		if tt.current != nil {
			i.Reviewers = tt.current
			i.CodeOwners = tt.currentCodeOwners
		}
		if tt.wantErr == nil {
			m.EXPECT().SetReviewers(gomock.Eq(ctx), gomock.Eq(i.Number), gomock.Eq(tt.want)).Return(nil)
		}
		if err := r.PerformReviewersAction(ctx, i, tt.in); err != nil {
			if !errors.As(err, tt.wantErr) {
				t.Errorf("got %v\nwant %v", err, tt.wantErr)
			}
		}
	}
}

func TestPerformCommentAction(t *testing.T) {
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

	m := mock.NewMockGhClient(ctrl)
	r.github = m

	tests := []struct {
		in           string
		sig          string
		mentionsEnv  string
		current      string
		want         string
		wantMentions []string
		wantErr      interface{}
	}{
		{"hello", "<!-- ghdag:test:sig -->", "", "", "hello\n<!-- ghdag:test:sig -->\n", []string{}, nil},
		{"hello", "<!-- ghdag:test:sig -->", "alice @bob", "", "hello\n<!-- ghdag:test:sig -->\n", []string{"alice", "@bob"}, nil},
		{"hello", "<!-- ghdag:test:sig -->", "", "hello\n<!-- ghdag:test:sig -->\n", "", []string{}, &erro.AlreadyInStateError{}},
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
		i.NumberOfConsecutiveComments = 1
		i.LatestCommentBody = tt.current
		if err := os.Setenv("GITHUB_COMMENT_MENTIONS", tt.mentionsEnv); err != nil {
			t.Fatal(err)
		}
		if tt.wantErr == nil {
			m.EXPECT().AddComment(gomock.Eq(ctx), gomock.Eq(i.Number), gomock.Eq(tt.want), gomock.Eq(tt.wantMentions)).Return(nil)
		}
		if err := r.PerformCommentAction(ctx, i, tt.in, tt.sig); err != nil {
			if !errors.As(err, tt.wantErr) {
				t.Errorf("got %v\nwant %v", err, tt.wantErr)
			}
		}
	}
}

func TestPerformStateAction(t *testing.T) {
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

	m := mock.NewMockGhClient(ctrl)
	r.github = m

	tests := []struct {
		in      string
		wantErr bool
	}{
		{"close", false},
		{"merge", false},
		{"revert", true},
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
		switch tt.in {
		case "close", "closed":
			m.EXPECT().CloseIssue(gomock.Eq(ctx), gomock.Eq(i.Number)).Return(nil)
			if err := r.PerformStateAction(ctx, i, tt.in); err != nil {
				t.Error(err)
			}
		case "merge", "merged":
			m.EXPECT().MergePullRequest(gomock.Eq(ctx), gomock.Eq(i.Number)).Return(nil)
			if err := r.PerformStateAction(ctx, i, tt.in); err != nil {
				t.Error(err)
			}
		default:
			if err := r.PerformStateAction(ctx, i, tt.in); (err != nil) != tt.wantErr {
				t.Errorf("got %v\nwant %v", err, tt.wantErr)
			}
		}
	}
}

func TestPerformNotifyAction(t *testing.T) {
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

	m := mock.NewMockSlkClient(ctrl)
	r.slack = m

	tests := []struct {
		in           string
		mentionsEnv  string
		want         string
		wantMentions []string
		wantErr      interface{}
	}{
		{"hello", "", "hello", []string{}, nil},
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
		if err := os.Setenv("SLACK_API_TOKEN", "dummy"); err != nil {
			t.Fatal(err)
		}
		if err := os.Setenv("SLACK_MENTIONS", tt.mentionsEnv); err != nil {
			t.Fatal(err)
		}
		if tt.wantErr == nil {
			m.EXPECT().PostMessage(gomock.Eq(ctx), gomock.Eq(tt.want), gomock.Eq(tt.wantMentions)).Return(nil)
		}
		if err := r.PerformNotifyAction(ctx, i, tt.in); err != nil {
			t.Error(err)
		}
	}
}
