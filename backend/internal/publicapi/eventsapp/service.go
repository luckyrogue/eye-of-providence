package eventsapp

import (
	"context"

	"github.com/eye-of-providence/backend/internal/store"
)

// Reader — недавние события пользователя.
type Reader interface {
	ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error)
}

// Service — GET /v1/public/events.
type Service struct{ r Reader }

// New — конструктор.
func New(r Reader) *Service { return &Service{r: r} }

// ListRecent — делегирует store.
func (s *Service) ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error) {
	if s.r == nil {
		return []store.Event{}, nil
	}
	return s.r.ListRecent(ctx, userID, limit)
}
