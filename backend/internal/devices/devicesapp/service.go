package devicesapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/devices/domain"
)

type Service struct {
	repo DeviceRepository
}

type Deps struct {
	Repo DeviceRepository
}

func New(d Deps) *Service {
	return &Service{repo: d.Repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]domain.Device, error) {
	if s.repo == nil {
		return []domain.Device{}, nil
	}
	return s.repo.List(ctx, userID)
}

func (s *Service) Revoke(ctx context.Context, userID, deviceID uuid.UUID) (bool, error) {
	if s.repo == nil {
		return false, nil
	}
	return s.repo.Revoke(ctx, userID, deviceID)
}
