package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth/meapp"
)

type pgMeProfileWriter struct {
	pool *pgxpool.Pool
}

func (p pgMeProfileWriter) UpdateName(ctx context.Context, userID uuid.UUID, displayName, lastName *string) error {
	if p.pool == nil {
		return nil
	}
	setExpr := []string{}
	args := []any{}
	if displayName != nil {
		args = append(args, *displayName)
		setExpr = append(setExpr, "display_name = $"+itoa(len(args)))
	}
	if lastName != nil {
		if *lastName == "" {
			args = append(args, nil)
		} else {
			args = append(args, *lastName)
		}
		setExpr = append(setExpr, "last_name = $"+itoa(len(args)))
	}
	args = append(args, userID)
	query := "UPDATE users SET " + strings.Join(setExpr, ", ") + " WHERE id = $" + itoa(len(args))
	_, err := p.pool.Exec(ctx, query, args...)
	return err
}

func (p pgMeProfileWriter) PasswordHash(ctx context.Context, userID uuid.UUID) (string, bool, error) {
	if p.pool == nil {
		return "", false, meapp.ErrDBNotConfigured
	}
	var hash *string
	err := p.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&hash)
	if err != nil {
		return "", false, err
	}
	if hash == nil || *hash == "" {
		return "", false, nil
	}
	return *hash, true, nil
}

func (p pgMeProfileWriter) UpdateEmail(ctx context.Context, userID uuid.UUID, email, _ string) error {
	if p.pool == nil {
		return meapp.ErrDBNotConfigured
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx,
		`UPDATE users SET email = $1, token_version = token_version + 1 WHERE id = $2`,
		email, userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return meapp.ErrEmailTaken
		}
		return err
	}
	return tx.Commit(ctx)
}

func (p pgMeProfileWriter) UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	if p.pool == nil {
		return meapp.ErrDBNotConfigured
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx,
		`UPDATE users SET password_hash = $1, token_version = token_version + 1 WHERE id = $2`,
		newHash, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type meSessionIssuer struct {
	pool      *pgxpool.Pool
	jwtSecret string
}

func (m meSessionIssuer) IssueAfterCredentialChange(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	tv, _ := TokenVersion(ctx, m.pool, userID)
	return IssueJWT(m.jwtSecret, userID.String(), email, "password", tv, defaultMeTokenTTL)
}

const defaultMeTokenTTL = 14 * 24 * time.Hour

func itoa(i int) string { return strconv.Itoa(i) }
