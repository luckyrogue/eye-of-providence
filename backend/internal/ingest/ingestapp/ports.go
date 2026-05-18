package ingestapp

import (
	"context"

	"github.com/eye-of-providence/backend/internal/ingest/domain"
)

type EventWriter interface {
	Insert(ctx context.Context, events []domain.Event) error
}
