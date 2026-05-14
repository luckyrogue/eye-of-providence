package prcomment

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentBody struct {
	Pool *pgxpool.Pool
	Base string
}

func (cb *CommentBody) Markdown(ctx context.Context, userID uuid.UUID, shas []string) (string, Aggregate, error) {
	agg, err := AggregateBySHA(ctx, cb.Pool, userID, shas)
	if err != nil {
		return "", agg, err
	}
	return FormatComment(agg, cb.Base), agg, nil
}
