package teams

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

type changeNameReq struct {
	DisplayName *string `json:"display_name"`
	LastName    *string `json:"last_name"`
}

func (s Service) handleChangeMyName(c *fiber.Ctx) error {
	if s.Pool == nil {
		return httperr.Unavailable(c, "db_required", "auth requires postgres")
	}
	var req changeNameReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	if req.DisplayName == nil && req.LastName == nil {
		return httperr.BadRequest(c, "no_fields", "provide display_name or last_name")
	}

	setExpr := []string{}
	args := []any{}
	if req.DisplayName != nil {
		dn, ok := validateDisplayName(*req.DisplayName)
		if !ok {
			return httperr.BadRequest(c, "invalid_display_name", "display_name 1..64 chars, no newlines")
		}
		args = append(args, dn)
		setExpr = append(setExpr, "display_name = $"+itoa(len(args)))
	}
	if req.LastName != nil {
		trimmed := strings.TrimSpace(*req.LastName)
		if trimmed == "" {
			args = append(args, nil)
			setExpr = append(setExpr, "last_name = $"+itoa(len(args)))
		} else {
			ln, ok := validateDisplayName(trimmed)
			if !ok {
				return httperr.BadRequest(c, "invalid_last_name", "last_name up to 64 chars, no newlines")
			}
			args = append(args, ln)
			setExpr = append(setExpr, "last_name = $"+itoa(len(args)))
		}
	}

	args = append(args, userID(c))
	query := "UPDATE users SET " + strings.Join(setExpr, ", ") + " WHERE id = $" + itoa(len(args))
	if _, err := s.Pool.Exec(c.Context(), query, args...); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

type changeEmailReq struct {
	CurrentPassword string `json:"current_password"`
	Email           string `json:"email"`
}

func (s Service) handleChangeMyEmail(c *fiber.Ctx) error {
	if s.Pool == nil {
		return httperr.Unavailable(c, "db_required", "auth requires postgres")
	}
	var req changeEmailReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	email, ok := validateEmail(req.Email)
	if !ok {
		return httperr.BadRequest(c, "invalid_email", "valid email required")
	}

	uid := userID(c)
	var hash *string
	if err := s.Pool.QueryRow(c.Context(),
		`SELECT password_hash FROM users WHERE id = $1`, uid).Scan(&hash); err != nil {
		return s.internalErr(c, err)
	}
	if hash == nil || *hash == "" {
		return httperr.BadRequest(c, "no_password_set", "set a password first (account linked via OAuth)")
	}
	if !auth.VerifyPassword(*hash, req.CurrentPassword) {
		return httperr.Unauthorized(c, "invalid_credentials", "incorrect password")
	}

	tx, err := s.Pool.Begin(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	defer func() { _ = tx.Rollback(c.Context()) }()
	if _, err := tx.Exec(c.Context(),
		`UPDATE users SET email = $1, token_version = token_version + 1 WHERE id = $2`,
		email, uid); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return httperr.Conflict(c, "email_taken", "email already in use")
		}
		return s.internalErr(c, err)
	}
	if err := tx.Commit(c.Context()); err != nil {
		return s.internalErr(c, err)
	}

	tv, _ := auth.TokenVersion(c.Context(), s.Pool, uid)
	tok, err := auth.IssueJWT(s.JWTSecret, uid.String(), email, "password", tv, tokenTTL)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"token": tok, "email": email})
}

type changePasswordReq struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s Service) handleChangeMyPassword(c *fiber.Ctx) error {
	if s.Pool == nil {
		return httperr.Unavailable(c, "db_required", "auth requires postgres")
	}
	var req changePasswordReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	if !validatePassword(req.NewPassword) {
		return httperr.BadRequest(c, "invalid_password", "password must be 8..256 chars")
	}

	uid := userID(c)
	var hash *string
	var email string
	if err := s.Pool.QueryRow(c.Context(),
		`SELECT password_hash, email FROM users WHERE id = $1`, uid).Scan(&hash, &email); err != nil {
		return s.internalErr(c, err)
	}
	if hash == nil || *hash == "" {
		return httperr.BadRequest(c, "no_password_set", "set a password first (account linked via OAuth)")
	}
	if !auth.VerifyPassword(*hash, req.CurrentPassword) {
		return httperr.Unauthorized(c, "invalid_credentials", "incorrect current password")
	}

	newHash, err := auth.HashPassword(req.NewPassword)
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

	tv, _ := auth.TokenVersion(c.Context(), s.Pool, uid)
	tok, err := auth.IssueJWT(s.JWTSecret, uid.String(), email, "password", tv, tokenTTL)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"token": tok})
}
