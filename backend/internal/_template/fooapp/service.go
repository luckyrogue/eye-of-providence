package fooapp

import (
	"context"

	"github.com/eye-of-providence/backend/internal/_template/domain"
)

type Service struct {
	repo domain.Repository
}

type Deps struct {
	Repo domain.Repository
}

func New(d Deps) *Service {
	return &Service{repo: d.Repo}
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Entity, error) {
	if s.repo == nil {
		return nil, domain.ErrNotFound
	}
	return s.repo.FindByID(ctx, id)
}
