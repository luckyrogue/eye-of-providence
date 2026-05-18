package prcomment

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/prcomment/prcommentapp"
)

type formatterAdapter struct {
	pool *pgxpool.Pool
	base string
}

func (f formatterAdapter) Markdown(ctx context.Context, userID uuid.UUID, shas []string) (string, prcommentapp.Aggregate, error) {
	comment, agg, err := (&CommentBody{Pool: f.pool, Base: f.base}).Markdown(ctx, userID, shas)
	return comment, prcommentapp.Aggregate{
		TotalCommits: agg.TotalCommits, WithAttribution: agg.WithAttribution,
		LinesAdded: agg.LinesAdded, LinesRemoved: agg.LinesRemoved,
		AIPercentWeighted: agg.AIPercentWeighted,
	}, err
}

type posterAdapter struct {
	hc HTTPClient
}

func (p posterAdapter) Post(ctx context.Context, provider, host, repo string, prNumber int, token, body string) error {
	return postComment(ctx, p.hc, request{
		Provider: provider, Host: host, Repo: repo, PRNumber: prNumber, ProviderToken: token,
	}, body)
}

func newPRCommentApp(pool *pgxpool.Pool, base string, hc HTTPClient) *prcommentapp.Service {
	return prcommentapp.New(prcommentapp.Deps{
		Format: formatterAdapter{pool: pool, base: base},
		Post:   posterAdapter{hc: hc},
	})
}
