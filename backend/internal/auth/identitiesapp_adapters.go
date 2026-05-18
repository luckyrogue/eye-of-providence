package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth/identitiesapp"
)

func newIdentitiesApp(pool *pgxpool.Pool) *identitiesapp.Service {
	if pool == nil {
		return identitiesapp.New(identitiesapp.Deps{})
	}
	return identitiesapp.New(identitiesapp.Deps{
		Repo:    pgIdentityRepo{pool: pool},
		Factors: authFactorCounterAdapter{pool: pool},
	})
}

type pgIdentityRepo struct {
	pool *pgxpool.Pool
}

func (p pgIdentityRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]identitiesapp.IdentityRow, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, provider, subject, COALESCE(email, ''), created_at
		FROM user_identities
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []identitiesapp.IdentityRow
	for rows.Next() {
		var r identitiesapp.IdentityRow
		if err := rows.Scan(&r.ID, &r.Provider, &r.Subject, &r.Email, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (p pgIdentityRepo) Delete(ctx context.Context, userID, identityID uuid.UUID) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`DELETE FROM user_identities WHERE id = $1 AND user_id = $2`, identityID, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

type authFactorCounterAdapter struct {
	pool *pgxpool.Pool
}

func (a authFactorCounterAdapter) Count(ctx context.Context, userID uuid.UUID, excludeIdentity *uuid.UUID, excludePasskey []byte) (int, error) {
	f, err := CountAuthFactors(ctx, a.pool, userID, excludeIdentity, excludePasskey)
	if err != nil {
		return 0, err
	}
	return f.Total(), nil
}
