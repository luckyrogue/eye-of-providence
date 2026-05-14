package prcomment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/prcomment"
)

func TestCommentBody_Markdown_EmptySHAs(t *testing.T) {
	cb := &prcomment.CommentBody{Pool: nil, Base: "https://dash"}
	md, agg, err := cb.Markdown(context.Background(), uuid.New(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if agg.TotalCommits != 0 || md == "" {
		t.Fatalf("md=%q agg=%+v", md, agg)
	}
}
