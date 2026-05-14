package devicelist

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DeviceRow — подмножество api_tokens для UI.
type DeviceRow struct {
	ID         uuid.UUID  `json:"id"`
	Kind       string     `json:"kind"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// Store — чтение списка устройств пользователя.
type Store interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]DeviceRow, error)
}

// Service — GET /v1/me/devices.
type Service struct {
	store Store
}

// New — конструктор.
func New(store Store) *Service {
	return &Service{store: store}
}

// ListMyDevices — список устройств (JWT user).
func (s *Service) ListMyDevices(ctx context.Context, userID uuid.UUID) ([]DeviceRow, error) {
	if s.store == nil {
		return []DeviceRow{}, nil
	}
	return s.store.ListByUser(ctx, userID)
}
