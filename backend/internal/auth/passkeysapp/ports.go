package passkeysapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PasskeyRow struct {
	ID         uuid.UUID  `json:"id"`
	Nickname   string     `json:"nickname,omitempty"`
	AAGUID     string     `json:"aaguid,omitempty"`
	Transports []string   `json:"transports,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type PasskeyRP interface {
	ListPasskeys(ctx context.Context, userID uuid.UUID) ([]PasskeyRow, error)
	PasskeyCredentialIDForUser(ctx context.Context, userID, passkeyID uuid.UUID) ([]byte, error)
	DeletePasskey(ctx context.Context, userID, passkeyID uuid.UUID) error
}

type AuthFactorCounter interface {
	Count(ctx context.Context, userID uuid.UUID, excludeIdentity *uuid.UUID, excludePasskey []byte) (int, error)
}
