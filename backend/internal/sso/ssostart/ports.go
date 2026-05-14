package ssostart

import (
	"context"

	"github.com/google/uuid"
)

type OIDCProvider interface {
	AuthCodeURL(stateValue, nonce string) string
}

type Registry interface {
	GetOIDC(ctx context.Context, teamID uuid.UUID) (OIDCProvider, error)
}

type StateCreator interface {
	CreateState(ctx context.Context, teamID uuid.UUID, returnTo string) (stateValue, nonce string, err error)
}
