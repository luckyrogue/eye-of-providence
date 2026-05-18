package pairingapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/devices/domain"
)

type Service struct {
	repo PairingRepository
}

type Deps struct {
	Repo PairingRepository
}

func New(d Deps) *Service {
	return &Service{repo: d.Repo}
}

func (s *Service) Begin(ctx context.Context, kind, nameHint string) (domain.PairBeginResult, error) {
	if s.repo == nil {
		return domain.PairBeginResult{}, nil
	}
	return s.repo.Begin(ctx, kind, nameHint)
}

func (s *Service) Poll(ctx context.Context, pairID uuid.UUID, secret string) (domain.PollResult, error) {
	if s.repo == nil {
		return domain.PollResult{}, nil
	}
	return s.repo.Poll(ctx, pairID, secret)
}

func (s *Service) Claim(ctx context.Context, userID uuid.UUID, code, name string) (domain.Device, error) {
	if s.repo == nil {
		return domain.Device{}, nil
	}
	return s.repo.Claim(ctx, userID, code, name)
}
