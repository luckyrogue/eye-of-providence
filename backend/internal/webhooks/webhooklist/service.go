package webhooklist

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Row struct {
	ID             uuid.UUID  `json:"id"`
	URL            string     `json:"url"`
	Events         []string   `json:"events"`
	Format         string     `json:"format"`
	Active         bool       `json:"active"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	LastStatus     *int       `json:"last_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Reader interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Row, error)
}

type Service struct {
	r Reader
}

func New(r Reader) *Service {
	return &Service{r: r}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Row, error) {
	if s.r == nil {
		return []Row{}, nil
	}
	return s.r.ListByUser(ctx, userID)
}
