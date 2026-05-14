package pushlist

import (
	"context"

	"github.com/google/uuid"
)

type SubscriptionRow struct {
	ID         uuid.UUID `json:"id"`
	Endpoint   string    `json:"endpoint"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  string    `json:"created_at"`
	LastUsedAt *string   `json:"last_used_at,omitempty"`
}

type Reader interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]SubscriptionRow, error)
}

type Service struct {
	r Reader
}

func New(r Reader) *Service {
	return &Service{r: r}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]SubscriptionRow, error) {
	if s.r == nil {
		return []SubscriptionRow{}, nil
	}
	return s.r.ListByUser(ctx, userID)
}
