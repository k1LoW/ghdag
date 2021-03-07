package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/k1LoW/ghdag/erro"
	"github.com/k1LoW/ghdag/mock"
	"github.com/k1LoW/ghdag/target"
)

func TestPerformLabelsAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
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
		r.revertEnv()
		ctx := context.Background()
		i := target.Target{}
		if err := faker.FakeData(&i); err != nil {
			t.Fatal(err)
		}
		if tt.current != nil {
			i.Labels = tt.current
		}
		if tt.wantErr == nil {
			m.EXPECT().SetLabels(gomock.Eq(ctx), gomock.Eq(i.Number), gomock.Eq(tt.want)).Return(nil)
			if err := r.PerformLabelsAction(ctx, &i, tt.in); err != nil {
				t.Error(err)
			}
		} else {
			if err := r.PerformLabelsAction(ctx, &i, tt.in); !errors.As(err, tt.wantErr) {
				t.Errorf("got %v\nwant %v", err, tt.wantErr)
			}
		}
	}
}
