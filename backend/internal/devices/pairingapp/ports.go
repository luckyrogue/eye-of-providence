package pairingapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/devices/domain"
)

type PairingRepository interface {
	Begin(ctx context.Context, kind, nameHint string) (domain.PairBeginResult, error)
	Poll(ctx context.Context, pairID uuid.UUID, secret string) (domain.PollResult, error)
	Claim(ctx context.Context, userID uuid.UUID, code, name string) (domain.Device, error)
}
