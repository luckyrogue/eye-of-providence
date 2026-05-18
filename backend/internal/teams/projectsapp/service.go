package projectsapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type Service struct {
	repo ProjectRepository
}

type Deps struct {
	Repo ProjectRepository
}

func New(d Deps) *Service {
	return &Service{repo: d.Repo}
}

func (s *Service) List(ctx context.Context, teamID uuid.UUID) ([]domain.Project, error) {
	if s.repo == nil {
		return []domain.Project{}, nil
	}
	return s.repo.List(ctx, teamID)
}

func (s *Service) Create(ctx context.Context, teamID, userID uuid.UUID, in domain.CreateProjectInput) (uuid.UUID, error) {
	if s.repo == nil {
		return uuid.Nil, nil
	}
	return s.repo.Create(ctx, teamID, userID, in)
}
