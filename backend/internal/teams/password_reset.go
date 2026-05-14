package teams

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/mailer"
)

const resetTokenTTL = 1 * time.Hour

func (s Service) handleForgotPassword(c *fiber.Ctx) error {
	if s.Pool == nil {

		return c.JSON(fiber.Map{"status": "ok"})
	}
	var req struct {
		Email string `json:"email"`
	}
	_ = c.BodyParser(&req)
	email, ok := validateEmail(req.Email)
	if !ok {
		return c.JSON(fiber.Map{"status": "ok"})
	}

	user, err := auth.FindUserByEmail(c.Context(), s.Pool, email)
	if errors.Is(err, auth.ErrUserNotFound) {
		return c.JSON(fiber.Map{"status": "ok"})
	}
	if err != nil {
		s.Logger.Warn("forgot-password lookup failed", zap.Error(err))
		return c.JSON(fiber.Map{"status": "ok"})
	}

	tok, hash, err := newResetToken()
	if err != nil {
		return s.internalErr(c, err)
	}
	expires := time.Now().Add(resetTokenTTL)
	if _, err := s.Pool.Exec(c.Context(), `
		INSERT INTO password_resets (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)`, hash, user.ID, expires); err != nil {
		s.Logger.Warn("forgot-password insert failed", zap.Error(err))
		return c.JSON(fiber.Map{"status": "ok"})
	}

	if s.Mailer != nil {

		var userLocale *string
		_ = s.Pool.QueryRow(c.Context(), `SELECT locale FROM users WHERE id = $1`, user.ID).Scan(&userLocale)
		loc := mailer.Locale("")
		if userLocale != nil {
			loc = mailer.Locale(*userLocale)
		}
		resetURL := strings.TrimRight(s.PublicURL, "/") + "/reset-password?token=" + url.QueryEscape(tok)
		subject, html, text := mailer.PasswordResetEmail(resetURL, loc)
		if err := s.Mailer.Send(c.Context(), email, subject, html, text); err != nil {
			s.Logger.Warn("forgot-password email send failed", zap.String("to", email), zap.Error(err))

		}
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func (s Service) handleResetPassword(c *fiber.Ctx) error {
	if s.Pool == nil {
		return httperr.Unavailable(c, "db_required", "auth requires postgres")
	}
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	if !validatePassword(req.Password) {
		return httperr.BadRequest(c, "invalid_password", "password must be 8..256 chars")
	}
	if req.Token == "" {
		return httperr.BadRequest(c, "missing_token", "missing token")
	}

	hash := hashResetToken(req.Token)
	uid, err := consumeResetToken(c.Context(), s.Pool, hash)
	if err != nil {

		return httperr.BadRequest(c, "reset_token_invalid", "invalid or expired reset token")
	}

	newHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return s.internalErr(c, err)
	}

	tx, err := s.Pool.Begin(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	defer func() { _ = tx.Rollback(c.Context()) }()
	if _, err := tx.Exec(c.Context(),
		`UPDATE users SET password_hash = $1, token_version = token_version + 1 WHERE id = $2`,
		newHash, uid); err != nil {
		return s.internalErr(c, err)
	}
	if err := tx.Commit(c.Context()); err != nil {
		return s.internalErr(c, err)
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func newResetToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	token = hex.EncodeToString(b)
	hash = hashResetToken(token)
	return
}

func hashResetToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func consumeResetToken(ctx context.Context, pool *pgxpool.Pool, tokenHash string) (uuid.UUID, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var uid uuid.UUID
	var expiresAt time.Time
	var usedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id, expires_at, used_at FROM password_resets
		WHERE token_hash = $1 FOR UPDATE
	`, tokenHash).Scan(&uid, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, errors.New("token not found")
	}
	if err != nil {
		return uuid.Nil, err
	}
	if usedAt != nil {
		return uuid.Nil, errors.New("token already used")
	}
	if time.Now().After(expiresAt) {
		return uuid.Nil, errors.New("token expired")
	}
	if _, err := tx.Exec(ctx, `UPDATE password_resets SET used_at = now() WHERE token_hash = $1`, tokenHash); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}
