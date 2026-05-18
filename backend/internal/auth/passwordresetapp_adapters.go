package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth/passwordresetapp"
	"github.com/eye-of-providence/backend/internal/mailer"
)

type resetTokenGen struct{}

func (resetTokenGen) NewToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	tok := hex.EncodeToString(b)
	return tok, resetTokenGen{}.HashToken(tok), nil
}

func (resetTokenGen) HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

type pgResetUsers struct{ pool *pgxpool.Pool }

func (p pgResetUsers) FindByEmail(ctx context.Context, email string) (uuid.UUID, *string, bool, error) {
	if p.pool == nil {
		return uuid.Nil, nil, false, nil
	}
	var id uuid.UUID
	var locale *string
	err := p.pool.QueryRow(ctx, `SELECT id, locale FROM users WHERE email = $1`, email).Scan(&id, &locale)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil, false, nil
	}
	if err != nil {
		return uuid.Nil, nil, false, err
	}
	return id, locale, true, nil
}

type pgResetTokens struct{ pool *pgxpool.Pool }

func (p pgResetTokens) Insert(ctx context.Context, userID uuid.UUID, tokenHash string, expires time.Time) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO password_resets (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)`, tokenHash, userID, expires)
	return err
}

func (p pgResetTokens) Consume(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	var uid uuid.UUID
	var expiresAt time.Time
	var usedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id, expires_at, used_at FROM password_resets
		WHERE token_hash = $1 FOR UPDATE`, tokenHash).Scan(&uid, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, passwordresetapp.ErrTokenInvalid
	}
	if err != nil {
		return uuid.Nil, err
	}
	if usedAt != nil || time.Now().After(expiresAt) {
		return uuid.Nil, passwordresetapp.ErrTokenInvalid
	}
	if _, err := tx.Exec(ctx, `UPDATE password_resets SET used_at = now() WHERE token_hash = $1`, tokenHash); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

type pgResetPassword struct{ pool *pgxpool.Pool }

func (p pgResetPassword) SetPassword(ctx context.Context, userID uuid.UUID, hash string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx,
		`UPDATE users SET password_hash = $1, token_version = token_version + 1 WHERE id = $2`,
		hash, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type resetMailAdapter struct {
	mailer mailer.Mailer
}

func (a resetMailAdapter) SendReset(ctx context.Context, to, resetURL, locale string) error {
	if a.mailer == nil {
		return nil
	}
	loc := mailer.Locale(locale)
	subject, html, text := mailer.PasswordResetEmail(resetURL, loc)
	return a.mailer.Send(ctx, to, subject, html, text)
}

type PasswordResetService struct {
	Pool      *pgxpool.Pool
	Mailer    mailer.Mailer
	PublicURL string
	Logger    *zap.Logger
}

func newPasswordResetApp(s PasswordResetService) *passwordresetapp.Service {
	return passwordresetapp.New(passwordresetapp.Deps{
		Users:     pgResetUsers{pool: s.Pool},
		Tokens:    pgResetTokens{pool: s.Pool},
		Mail:      resetMailAdapter{mailer: s.Mailer},
		Password:  pgResetPassword{pool: s.Pool},
		TokensGen: resetTokenGen{},
		PublicURL: s.PublicURL,
	})
}
