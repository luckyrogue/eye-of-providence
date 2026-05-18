package pushapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/push/domain"
)

type Service struct {
	repo Repository
}

type Deps struct {
	Repo Repository
}

func New(d Deps) *Service {
	return &Service{repo: d.Repo}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	if s.repo == nil {
		return []domain.Subscription{}, nil
	}
	return s.repo.List(ctx, userID)
}

func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth, userAgent string) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.Subscribe(ctx, userID, endpoint, p256dh, auth, userAgent)
}

func (s *Service) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) (bool, error) {
	if s.repo == nil {
		return false, nil
	}
	return s.repo.Unsubscribe(ctx, userID, endpoint)
}
