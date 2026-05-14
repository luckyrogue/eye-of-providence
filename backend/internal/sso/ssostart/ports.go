package ssostart

import (
	"context"

	"github.com/google/uuid"
)

// OIDCProvider — минимум для authorize URL (реализация: *sso.OIDCProvider).
type OIDCProvider interface {
	AuthCodeURL(stateValue, nonce string) string
}

// Registry — резолвинг OIDC-провайдера по команде.
type Registry interface {
	GetOIDC(ctx context.Context, teamID uuid.UUID) (OIDCProvider, error)
}

// StateCreator — CSRF state + nonce в БД.
type StateCreator interface {
	CreateState(ctx context.Context, teamID uuid.UUID, returnTo string) (stateValue, nonce string, err error)
}
