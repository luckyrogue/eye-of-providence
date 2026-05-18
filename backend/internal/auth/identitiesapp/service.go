package identitiesapp

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo     IdentityRepository
	factors  AuthFactorCounter
}

type Deps struct {
	Repo    IdentityRepository
	Factors AuthFactorCounter
}

func New(d Deps) *Service {
	return &Service{repo: d.Repo, factors: d.Factors}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]IdentityRow, error) {
	if s.repo == nil {
		return []IdentityRow{}, nil
	}
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, userID, identityID uuid.UUID) error {
	if s.repo == nil {
		return ErrDBRequired
	}
	if s.factors != nil {
		n, err := s.factors.Count(ctx, userID, &identityID, nil)
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrLastAuthFactor
		}
	}
	ok, err := s.repo.Delete(ctx, userID, identityID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}
