package teamsapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TeamRow возвращается из handler'ов через c.JSON, поэтому JSON-теги
// обязательны — без них Go использует имена полей as-is (ID, Name),
// а dashboard и интеграционные тесты ждут lowercase.
type TeamRow struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Role              string     `json:"role"`
	SubscriptionPlan  string     `json:"subscription_plan,omitempty"`
	SubscriptionUntil *time.Time `json:"subscription_until,omitempty"`
	SubscriptionNote  *string    `json:"subscription_note,omitempty"`
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
