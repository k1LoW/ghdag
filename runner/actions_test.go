package runner

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/k1LoW/ghdag/env"
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
		if got := os.Getenv("GHDAG_ACTION_LABELS_UPDATED"); got != env.Join(tt.want) {
			t.Errorf("got %v\nwant %v", got, env.Join(tt.want))
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
		if got := os.Getenv("GHDAG_ACTION_ASSIGNEES_UPDATED"); got != env.Join(tt.want) {
			t.Errorf("got %v\nwant %v", got, env.Join(tt.want))
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
		if got := os.Getenv("GHDAG_ACTION_REVIEWERS_UPDATED"); got != env.Join(tt.want) {
			t.Errorf("got %v\nwant %v", got, env.Join(tt.want))
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
		in          string
		sig         string
		mentionsEnv string
		current     string
		want        string
		wantErr     interface{}
	}{
		{"hello", "<!-- ghdag:test:sig -->", "", "", "hello\n<!-- ghdag:test:sig -->\n", nil},
		{"hello", "<!-- ghdag:test:sig -->", "alice @bob", "", "@alice @bob hello\n<!-- ghdag:test:sig -->\n", nil},
		{"hello", "<!-- ghdag:test:sig -->", "", "hello\n<!-- ghdag:test:sig -->\n", "", &erro.AlreadyInStateError{}},
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
			m.EXPECT().AddComment(gomock.Eq(ctx), gomock.Eq(i.Number), gomock.Eq(tt.want)).Return(nil)
		}
		if err := r.PerformCommentAction(ctx, i, tt.in, tt.sig); err != nil {
			if !errors.As(err, tt.wantErr) {
				t.Errorf("got %v\nwant %v", err, tt.wantErr)
			}
		}
		if got := os.Getenv("GHDAG_ACTION_COMMENT_CREATED"); got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
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
		want    string
		wantErr bool
	}{
		{"close", "closed", false},
		{"merge", "merged", false},
		{"revert", "", true},
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
		if got := os.Getenv("GHDAG_ACTION_STATE_CHANGED"); got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
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
		{"hello", "alice", "<UALICE> hello", []string{"alice"}, nil},
		{"hello", "alice bob", "<UALICE> <UBOB> hello", []string{"alice", "bob"}, nil},
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
			m.EXPECT().PostMessage(gomock.Eq(ctx), gomock.Eq(tt.want)).Return(nil)
			for _, mention := range tt.wantMentions {
				m.EXPECT().GetMentionLinkByName(gomock.Eq(ctx), gomock.Eq(mention)).Return(fmt.Sprintf("<U%s>", strings.ToUpper(mention)), nil)
			}
		}
		if err := r.PerformNotifyAction(ctx, i, tt.in); err != nil {
			t.Error(err)
		}
		if got := os.Getenv("GHDAG_ACTION_NOTIFY_SENT"); got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestSetReviewersAndNotify(t *testing.T) {
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
	mg := mock.NewMockGhClient(ctrl)
	ms := mock.NewMockSlkClient(ctrl)
	r.github = mg
	r.slack = ms

	tests := []struct {
		authorExist    bool
		enableSameSeed bool
	}{
		{false, true},
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
		if err := os.Setenv("GHDAG_SAMPLE_WITH_SAME_SEED", fmt.Sprintf("%t", tt.enableSameSeed)); err != nil {
			t.Fatal(err)
		}
		if err := os.Setenv("SLACK_API_TOKEN", "dummy"); err != nil {
			t.Fatal(err)
		}
		users := []string{}
		for i := 0; i < 100; i++ {
			users = append(users, fmt.Sprintf("user%d", i))
		}
		sample := rand.Intn(100)
		if err := os.Setenv("GITHUB_REVIEWERS_SAMPLE", fmt.Sprintf("%d", sample)); err != nil {
			t.Fatal(err)
		}
		if err := os.Setenv("SLACK_MENTIONS", env.Join(users)); err != nil {
			t.Fatal(err)
		}

		mg.EXPECT().SetReviewers(gomock.Eq(ctx), gomock.Eq(i.Number), gomock.Any()).Return(nil)
		r.initSeed()
		if err := r.PerformReviewersAction(ctx, i, users); err != nil {
			t.Errorf("got %v", err)
		}

		want, err := env.Split(os.Getenv("GHDAG_ACTION_REVIEWERS_UPDATED"))
		if err != nil {
			t.Fatal(err)
		}

		if sample != len(want) {
			t.Errorf("got %v\nwant %v", sample, len(want))
		}

		if err := os.Setenv("SLACK_MENTIONS_SAMPLE", fmt.Sprintf("%d", sample)); err != nil {
			t.Fatal(err)
		}
		ms.EXPECT().PostMessage(gomock.Eq(ctx), gomock.Any()).Return(nil)
		for _, mu := range want {
			ms.EXPECT().GetMentionLinkByName(gomock.Eq(ctx), gomock.Eq(mu)).Return("<UTEST>", nil)
		}
		r.initSeed()
		if err := r.PerformNotifyAction(ctx, i, "Hello"); err != nil {
			t.Errorf("got %v", err)
		}
	}
}
