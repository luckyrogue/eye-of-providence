package membersapp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type MemberRepository interface {
	ListByTeam(ctx context.Context, teamID domain.TeamID) ([]domain.Member, error)
}

type RoleReader interface {
	TeamRole(ctx context.Context, userID, teamID uuid.UUID) (role string, ok bool)
}

type MemberMutator interface {
	TargetRole(ctx context.Context, teamID, userID uuid.UUID) (string, error)
	OwnerCount(ctx context.Context, teamID uuid.UUID) (int, error)
	OwnedTeamCount(ctx context.Context, userID uuid.UUID, excludeTeamID uuid.UUID) (int, error)
	UpdateRole(ctx context.Context, teamID, userID uuid.UUID, role string) error
	Remove(ctx context.Context, teamID, userID uuid.UUID) error
}

type ActivityAggregator interface {
	AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error)
}
