package webhooksapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/webhooks/domain"
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

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]domain.Webhook, error) {
	if s.repo == nil {
		return []domain.Webhook{}, nil
	}
	return s.repo.List(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, url string, events []string, format string) (string, domain.Webhook, error) {
	if s.repo == nil {
		return "", domain.Webhook{}, nil
	}
	return s.repo.Create(ctx, userID, url, events, format)
}

func (s *Service) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	if s.repo == nil {
		return false, nil
	}
	return s.repo.Delete(ctx, userID, id)
}
