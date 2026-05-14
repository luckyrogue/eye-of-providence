package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth/passwordapp"
)

type pgPasswordLoginReader struct {
	pool *pgxpool.Pool
}

func (p pgPasswordLoginReader) LookupByEmail(ctx context.Context, email string) (emailOut, displayName, passwordHash string, userID uuid.UUID, err error) {
	if p.pool == nil {
		return "", "", "", uuid.Nil, passwordapp.ErrDBNotConfigured
	}
	u, err := FindUserByEmail(ctx, p.pool, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", "", "", uuid.Nil, passwordapp.ErrInvalidCredentials
		}
		return "", "", "", uuid.Nil, err
	}
	return u.Email, u.DisplayName, u.PasswordHash, u.ID, nil
}

func NewPasswordLoginService(pool *pgxpool.Pool) *passwordapp.Service {
	if pool == nil {
		return passwordapp.New(nil)
	}
	return passwordapp.New(pgPasswordLoginReader{pool: pool})
}
