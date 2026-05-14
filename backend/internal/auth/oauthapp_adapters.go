package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth/oauthapp"
)

type pgOAuthStore struct{ pool *pgxpool.Pool }

func (p pgOAuthStore) FindUserIDByIdentity(ctx context.Context, provider, subject string) (uuid.UUID, bool, error) {
	var uid uuid.UUID
	err := p.pool.QueryRow(ctx,
		`SELECT user_id FROM user_identities WHERE provider = $1 AND subject = $2`,
		provider, subject,
	).Scan(&uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return uid, true, nil
}

func (p pgOAuthStore) UpdateUserEmailIfEmpty(ctx context.Context, userID uuid.UUID, email string) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE users SET email = COALESCE(NULLIF(users.email, ''), $1) WHERE id = $2`,
		email, userID,
	)
	return err
}

func (p pgOAuthStore) FindUserIDByEmail(ctx context.Context, email string) (uuid.UUID, bool, error) {
	var uid uuid.UUID
	err := p.pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return uid, true, nil
}

func (p pgOAuthStore) LinkIdentity(ctx context.Context, userID uuid.UUID, provider, subject, email string) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO user_identities (user_id, provider, subject, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, subject) DO NOTHING
	`, userID, provider, subject, email)
	return err
}

func (p pgOAuthStore) CreateUserWithIdentity(ctx context.Context, newID uuid.UUID, provider string, ext oauthapp.ExternalUser) error {
	displayName := ext.Name
	if displayName == "" {
		displayName = ext.Login
	}
	if displayName == "" {
		displayName = ext.Email
	}
	githubLogin := ""
	if provider == "github" {
		githubLogin = ext.Login
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, github_login, display_name)
		VALUES ($1, $2, NULLIF($3, ''), $4)
		ON CONFLICT (id) DO NOTHING
	`, newID, ext.Email, githubLogin, displayName); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_identities (user_id, provider, subject, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, subject) DO NOTHING
	`, newID, provider, ext.Subject, ext.Email); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func newOAuthAppService(s Service) *oauthapp.Service {
	if s.Pool == nil {
		return oauthapp.New(nil)
	}
	return oauthapp.New(pgOAuthStore{pool: s.Pool})
}
