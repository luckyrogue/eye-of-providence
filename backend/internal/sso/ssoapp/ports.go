package ssoapp

import (
	"context"

	"github.com/google/uuid"
)

type OIDCProvider interface {
	AuthCodeURL(state, nonce string) string
}

type Registry interface {
	GetOIDC(ctx context.Context, teamID uuid.UUID) (OIDCProvider, error)
}

type StateCreator interface {
	CreateState(ctx context.Context, teamID uuid.UUID, returnTo string) (state, nonce string, err error)
}
