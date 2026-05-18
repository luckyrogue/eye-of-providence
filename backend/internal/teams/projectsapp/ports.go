package projectsapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type ProjectRepository interface {
	List(ctx context.Context, teamID uuid.UUID) ([]domain.Project, error)
	Create(ctx context.Context, teamID, userID uuid.UUID, in domain.CreateProjectInput) (uuid.UUID, error)
}
