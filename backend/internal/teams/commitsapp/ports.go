package commitsapp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type CommitIngestor interface {
	Ingest(ctx context.Context, userID uuid.UUID, in domain.CommitInput) (inserted bool, teamID uuid.UUID, projectID uuid.UUID, err error)
}

type CommitRow struct {
	ID           uuid.UUID
	ProjectID    *uuid.UUID
	UserID       uuid.UUID
	Author       string
	SHA          string
	Message      string
	Branch       string
	FilesChanged int
	LinesAdded   int
	LinesRemoved int
	AILinesPct   *int
	AuthoredAt   time.Time
}

type CommitReader interface {
	ListByProject(ctx context.Context, teamID, projectID uuid.UUID) ([]CommitRow, error)
	ListByTeam(ctx context.Context, teamID uuid.UUID) ([]CommitRow, error)
}

type ProjectTeamResolver interface {
	TeamIDForProject(ctx context.Context, projectID uuid.UUID) (*uuid.UUID, error)
}

type RoleReader interface {
	TeamRole(ctx context.Context, userID, teamID uuid.UUID) (role string, ok bool)
}
