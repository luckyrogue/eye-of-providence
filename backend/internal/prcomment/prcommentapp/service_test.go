package prcommentapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/prcomment/prcommentapp"
)

type fakeFmt struct{}

func (fakeFmt) Markdown(context.Context, uuid.UUID, []string) (string, prcommentapp.Aggregate, error) {
	return "# comment", prcommentapp.Aggregate{TotalCommits: 1}, nil
}

type fakePost struct{}

func (fakePost) Post(context.Context, string, string, string, int, string, string) error {
	return nil
}

func TestPostPRComment(t *testing.T) {
	svc := prcommentapp.New(prcommentapp.Deps{Format: fakeFmt{}, Post: fakePost{}})
	agg, md, err := svc.PostPRComment(context.Background(), uuid.New(), prcommentapp.PostRequest{
		Provider: "github", Repo: "o/r", PRNumber: 1, SHAs: []string{"abc"},
	})
	if err != nil || md == "" || agg.TotalCommits != 1 {
		t.Fatalf("agg=%+v md=%q err=%v", agg, md, err)
	}
}
