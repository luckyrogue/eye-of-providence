package teamsapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TeamRow struct {
	ID                uuid.UUID
	Name              string
	Role              string
	SubscriptionPlan  string
	SubscriptionUntil *time.Time
	SubscriptionNote  *string
}

type TeamRepository interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]TeamRow, error)
	GetName(ctx context.Context, teamID uuid.UUID) (string, error)
	Create(ctx context.Context, in CreateTeamParams) (uuid.UUID, error)
	UpdateName(ctx context.Context, teamID uuid.UUID, name string) error
	Delete(ctx context.Context, teamID uuid.UUID) error
}

type CreateTeamParams struct {
	UserID   uuid.UUID
	Name     string
	IsSuper  bool
	BetaLimit int
	LockID   int64
}

type BetaGate interface {
	TeamCount(ctx context.Context) (int, error)
}

type OwnerLimitChecker interface {
	OwnedTeamCount(ctx context.Context, userID uuid.UUID) (int, error)
}
