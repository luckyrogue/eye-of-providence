package eventsapp

import (
	"context"

	"github.com/eye-of-providence/backend/internal/store"
)

type Reader interface {
	ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error)
}

type Service struct{ r Reader }

func New(r Reader) *Service { return &Service{r: r} }

func (s *Service) ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error) {
	if s.r == nil {
		return []store.Event{}, nil
	}
	return s.r.ListRecent(ctx, userID, limit)
}
