package sso

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	TokenVersion int
}

var ErrJITDisabled = errors.New("user not found and JIT provisioning disabled")

func ProvisionUser(
	ctx context.Context,
	pool *pgxpool.Pool,
	teamID uuid.UUID,
	cfg *Config,
	ident *OIDCIdentity,
) (*User, bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var u User
	err = tx.QueryRow(ctx, `
		SELECT id, email, COALESCE(display_name, email), token_version
		FROM users WHERE sso_team_id = $1 AND sso_subject = $2`,
		teamID, ident.Subject,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.TokenVersion)
	if err == nil {

		if err := ensureTeamMember(ctx, tx, teamID, u.ID, cfg.JITRole); err != nil {
			return nil, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}
		return &u, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	err = tx.QueryRow(ctx, `
		SELECT id, email, COALESCE(display_name, email), token_version
		FROM users WHERE email = $1`,
		ident.Email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.TokenVersion)
	if err == nil {

		if _, err := tx.Exec(ctx, `
			UPDATE users SET sso_team_id = $1, sso_subject = $2 WHERE id = $3`,
			teamID, ident.Subject, u.ID,
		); err != nil {
			return nil, false, err
		}
		if err := ensureTeamMember(ctx, tx, teamID, u.ID, cfg.JITRole); err != nil {
			return nil, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}
		return &u, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	if !cfg.JITProvision {
		return nil, false, ErrJITDisabled
	}

	uid := uuid.New()
	displayName := ident.Name
	if displayName == "" {

		displayName = emailLocalPart(ident.Email)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, display_name, sso_team_id, sso_subject, token_version)
		VALUES ($1, $2, $3, $4, $5, 1)`,
		uid, ident.Email, displayName, teamID, ident.Subject,
	); err != nil {
		return nil, false, err
	}
	if err := ensureTeamMember(ctx, tx, teamID, uid, cfg.JITRole); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return &User{
		ID:           uid,
		Email:        ident.Email,
		DisplayName:  displayName,
		TokenVersion: 1,
	}, true, nil
}

func ensureTeamMember(ctx context.Context, tx pgx.Tx, teamID, userID uuid.UUID, role string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO NOTHING`,
		teamID, userID, role,
	)
	return err
}

func emailLocalPart(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return email
	}
	return email[:at]
}

func Touch(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, ident *OIDCIdentity) {
	if ident.Name == "" {
		_, _ = pool.Exec(ctx, `UPDATE users SET email = $1 WHERE id = $2`,
			ident.Email, userID)
		return
	}
	_, _ = pool.Exec(ctx, `UPDATE users SET email = $1, display_name = $2 WHERE id = $3`,
		ident.Email, ident.Name, userID)
}
