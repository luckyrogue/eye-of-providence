package domain

import "context"

// BlockRepository — persistence port для aggregate Block.
type BlockRepository interface {
	Lookup(ctx context.Context, slug, locale string, includeDraft bool) (*Block, error)
	Upsert(ctx context.Context, p UpsertParams) (*Block, error)
	Delete(ctx context.Context, slug, locale string) (*Block, error)
	ListMatrix(ctx context.Context) ([]MatrixEntry, error)
}
