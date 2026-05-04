package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UsersPG — лёгкий upsert users для dev-token и github oauth.
// Полная users-таблица расширяется в Phase 7 (teams, role, profile).
type UsersPG struct {
	pool *pgxpool.Pool
}

func NewUsersPG(pool *pgxpool.Pool) *UsersPG {
	return &UsersPG{pool: pool}
}

func (u *UsersPG) Upsert(ctx context.Context, userID uuid.UUID, email, githubLogin string) error {
	if u == nil || u.pool == nil {
		return nil
	}
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := u.pool.Exec(c, `
		INSERT INTO users (id, email, github_login)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''))
		ON CONFLICT (id) DO UPDATE SET
		  email = COALESCE(EXCLUDED.email, users.email),
		  github_login = COALESCE(EXCLUDED.github_login, users.github_login)
	`, userID, email, githubLogin)
	return err
}
