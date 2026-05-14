package sso

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/sso/ssostart"
)

type ssostartRegistry struct{ r *Registry }

func (w ssostartRegistry) GetOIDC(ctx context.Context, teamID uuid.UUID) (ssostart.OIDCProvider, error) {
	return w.r.Get(ctx, teamID)
}

type ssostartStates struct{ pool *pgxpool.Pool }

func (w ssostartStates) CreateState(ctx context.Context, teamID uuid.UUID, returnTo string) (string, string, error) {
	st, err := CreateState(ctx, w.pool, teamID, returnTo)
	if err != nil {
		return "", "", err
	}
	return st.Value, st.Nonce, nil
}

func newSSOStartService(s Service) *ssostart.Service {
	return ssostart.New(ssostartRegistry{r: s.Registry}, ssostartStates{pool: s.Pool})
}
