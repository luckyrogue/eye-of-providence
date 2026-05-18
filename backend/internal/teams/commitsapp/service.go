package commitsapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type Service struct {
	ingest   CommitIngestor
	commits  CommitReader
	projects ProjectTeamResolver
	roles    RoleReader
}

type Deps struct {
	Ingest   CommitIngestor
	Commits  CommitReader
	Projects ProjectTeamResolver
	Roles    RoleReader
}

func New(d Deps) *Service {
	return &Service{
		ingest: d.Ingest, commits: d.Commits, projects: d.Projects, roles: d.Roles,
	}
}

func (s *Service) Ingest(ctx context.Context, userID uuid.UUID, in domain.CommitInput) (bool, uuid.UUID, uuid.UUID, error) {
	if s.ingest == nil {
		return false, uuid.Nil, uuid.Nil, nil
	}
	return s.ingest.Ingest(ctx, userID, in)
}

func (s *Service) IngestForMember(ctx context.Context, userID uuid.UUID, in domain.CommitInput) (bool, error) {
	projID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return false, err
	}
	teamID, err := s.projects.TeamIDForProject(ctx, projID)
	if err != nil {
		return false, err
	}
	if teamID == nil {
		return false, domain.ErrProjectOrphaned
	}
	if s.roles != nil {
		if _, ok := s.roles.TeamRole(ctx, userID, *teamID); !ok {
			return false, domain.ErrNotMember
		}
	}
	inserted, _, _, err := s.ingest.Ingest(ctx, userID, in)
	return inserted, err
}

func (s *Service) ListByProject(ctx context.Context, userID, teamID, projectID uuid.UUID) ([]CommitRow, error) {
	if s.roles != nil {
		if _, ok := s.roles.TeamRole(ctx, userID, teamID); !ok {
			return nil, domain.ErrNotMember
		}
	}
	if s.commits == nil {
		return []CommitRow{}, nil
	}
	return s.commits.ListByProject(ctx, teamID, projectID)
}

func (s *Service) ListByTeam(ctx context.Context, userID, teamID uuid.UUID) ([]CommitRow, error) {
	if s.roles != nil {
		if _, ok := s.roles.TeamRole(ctx, userID, teamID); !ok {
			return nil, domain.ErrNotMember
		}
	}
	if s.commits == nil {
		return []CommitRow{}, nil
	}
	return s.commits.ListByTeam(ctx, teamID)
}
