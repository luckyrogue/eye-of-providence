package devicelist

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceRow struct {
	ID         uuid.UUID  `json:"id"`
	Kind       string     `json:"kind"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type Store interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]DeviceRow, error)
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListMyDevices(ctx context.Context, userID uuid.UUID) ([]DeviceRow, error) {
	if s.store == nil {
		return []DeviceRow{}, nil
	}
	return s.store.ListByUser(ctx, userID)
}
