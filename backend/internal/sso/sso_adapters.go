package sso

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/sso/ssoapp"
)

type ssoRegistryAdapter struct{ r *Registry }

func (w ssoRegistryAdapter) GetOIDC(ctx context.Context, teamID uuid.UUID) (ssoapp.OIDCProvider, error) {
	return w.r.Get(ctx, teamID)
}

type ssoStatesAdapter struct{ pool *pgxpool.Pool }

func (w ssoStatesAdapter) CreateState(ctx context.Context, teamID uuid.UUID, returnTo string) (string, string, error) {
	st, err := CreateState(ctx, w.pool, teamID, returnTo)
	if err != nil {
		return "", "", err
	}
	return st.Value, st.Nonce, nil
}

func newSSOApp(s Service) *ssoapp.Service {
	return ssoapp.New(ssoapp.Deps{
		Registry: ssoRegistryAdapter{r: s.Registry},
		States:   ssoStatesAdapter{pool: s.Pool},
	})
}
