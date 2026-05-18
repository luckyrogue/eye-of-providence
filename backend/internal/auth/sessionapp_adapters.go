package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth/sessionapp"
)

func newSessionApp(jwtSecret string, pool *pgxpool.Pool) *sessionapp.Service {
	return sessionapp.New(sessionapp.Deps{
		Signer: jwtSignerAdapter{secret: jwtSecret, pool: pool},
	})
}

type jwtSignerAdapter struct {
	secret string
	pool   *pgxpool.Pool
}

func (a jwtSignerAdapter) Issue(ctx context.Context, userID uuid.UUID, email, method string) (string, error) {
	if email == "" && a.pool != nil {
		_ = a.pool.QueryRow(ctx,
			`SELECT COALESCE(email, '') FROM users WHERE id = $1`, userID,
		).Scan(&email)
	}
	tv, _ := TokenVersion(ctx, a.pool, userID)
	return IssueJWT(a.secret, userID.String(), email, method, tv, tokenTTL)
}
