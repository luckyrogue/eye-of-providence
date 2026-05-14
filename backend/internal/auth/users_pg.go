package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func TokenVersion(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (int, error) {
	var tv int
	err := pool.QueryRow(ctx,
		"SELECT token_version FROM users WHERE id=$1", userID).Scan(&tv)
	return tv, err
}

func BumpTokenVersion(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	if pool == nil {
		return nil
	}
	_, err := pool.Exec(ctx,
		"UPDATE users SET token_version = token_version + 1 WHERE id=$1", userID)
	return err
}
