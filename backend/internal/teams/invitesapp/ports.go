package invitesapp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type InviteRepository interface {
	FindByCode(ctx context.Context, code string) (*domain.Invite, error)
	Consume(ctx context.Context, code string, userID uuid.UUID) (teamID uuid.UUID, err error)
	Create(ctx context.Context, teamID, createdBy uuid.UUID, code string, maxUses int, expires time.Time, email *string) error
	MarkSent(ctx context.Context, code string, sentAt time.Time) error
}

type MemberAdder interface {
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error
}

type TeamReader interface {
	Name(ctx context.Context, teamID uuid.UUID) (string, error)
	Plan(ctx context.Context, teamID uuid.UUID) (string, error)
	MemberCount(ctx context.Context, teamID uuid.UUID) (int, error)
}

type InviteMailer interface {
	Send(ctx context.Context, teamID, inviterID uuid.UUID, to, code string) error
}

type PlanLimits struct {
	Plan            string
	MaxUsersPerTeam int
}

type PlanLimiter interface {
	Limits(plan string) PlanLimits
}

type RoleChecker interface {
	TeamRole(ctx context.Context, userID, teamID uuid.UUID) (string, bool)
}
