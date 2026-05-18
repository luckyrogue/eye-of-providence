package identitiesapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type IdentityRow struct {
	ID        uuid.UUID `json:"id"`
	Provider  string    `json:"provider"`
	Subject   string    `json:"subject"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type IdentityRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]IdentityRow, error)
	Delete(ctx context.Context, userID, identityID uuid.UUID) (bool, error)
}

type AuthFactorCounter interface {
	Count(ctx context.Context, userID uuid.UUID, excludeIdentity *uuid.UUID, excludePasskey []byte) (int, error)
}
