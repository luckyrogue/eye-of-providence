package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthFactors struct {
	PasswordSet bool
	Identities  int
	Passkeys    int
}

func (f AuthFactors) Total() int {
	n := f.Identities + f.Passkeys
	if f.PasswordSet {
		n++
	}
	return n
}

func CountAuthFactors(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	excludeIdentity *uuid.UUID,
	excludePasskey []byte,
) (AuthFactors, error) {
	out := AuthFactors{}
	if pool == nil {
		return out, nil
	}
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var hasPassword bool
	if err := pool.QueryRow(c,
		`SELECT password_hash IS NOT NULL FROM users WHERE id = $1`, userID,
	).Scan(&hasPassword); err != nil {
		return out, err
	}
	out.PasswordSet = hasPassword

	idQuery := `SELECT count(*) FROM user_identities WHERE user_id = $1`
	idArgs := []any{userID}
	if excludeIdentity != nil {
		idQuery += ` AND id <> $2`
		idArgs = append(idArgs, *excludeIdentity)
	}
	if err := pool.QueryRow(c, idQuery, idArgs...).Scan(&out.Identities); err != nil {
		return out, err
	}

	pkQuery := `SELECT count(*) FROM webauthn_credentials WHERE user_id = $1`
	pkArgs := []any{userID}
	if len(excludePasskey) > 0 {
		pkQuery += ` AND credential_id <> $2`
		pkArgs = append(pkArgs, excludePasskey)
	}
	if err := pool.QueryRow(c, pkQuery, pkArgs...).Scan(&out.Passkeys); err != nil {
		return out, err
	}

	return out, nil
}
