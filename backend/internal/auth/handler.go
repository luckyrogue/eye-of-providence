package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth/oauthapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

const (
	tokenTTL = 14 * 24 * time.Hour

	handoffCookieName = "eop_session_handoff"
	handoffTTL        = 30 * time.Second
)

type Service struct {
	JWTSecret string

	Providers      map[string]OAuthProvider
	GitHub         *GitHubOAuth
	Logger         *zap.Logger
	Users          *UsersPG
	Pool           *pgxpool.Pool
	WebAuthn       *WebAuthnService
	PublicURL      string
	SecureCookies  bool
	EnableDevToken bool
}

func (s Service) ProvidersList() []string {
	out := make([]string, 0, len(s.Providers))
	for name := range s.Providers {
		out = append(out, name)
	}

	order := []string{"github", "google", "apple"}
	sorted := make([]string, 0, len(out))
	for _, k := range order {
		for _, n := range out {
			if n == k {
				sorted = append(sorted, n)
				break
			}
		}
	}

	for _, n := range out {
		found := false
		for _, k := range order {
			if n == k {
				found = true
				break
			}
		}
		if !found {
			sorted = append(sorted, n)
		}
	}
	return sorted
}

func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/auth")

	if s.EnableDevToken {
		registerDevToken(g, s)
	}

	for name, p := range s.Providers {
		registerOAuthProvider(g, s, name, p)
	}

	if _, has := s.Providers["github"]; !has && s.GitHub != nil {
		registerOAuthProvider(g, s, "github", s.GitHub)
	}

	g.Get("/session-handoff-internal", handleSessionHandoff(s))

	if s.WebAuthn != nil {

		g.Post("/webauthn/login/begin", handleWebAuthnLoginBegin(s))
		g.Post("/webauthn/login/finish", handleWebAuthnLoginFinish(s))

		mw := Middleware(s.JWTSecret, s.Pool)
		g.Post("/webauthn/register/begin", mw, handleWebAuthnRegisterBegin(s))
		g.Post("/webauthn/register/finish", mw, handleWebAuthnRegisterFinish(s))
	}
}

func RegisterSessionHandoffRoute(app *fiber.App, s Service) {
	app.Get("/v1/me/session-handoff", handleSessionHandoff(s))
}

func registerOAuthProvider(g fiber.Router, s Service, name string, p OAuthProvider) {
	g.Get("/"+name+"/login", func(c *fiber.Ctx) error {
		state := randomState()

		returnTo := c.Query("return_to")
		stateVal := state
		if returnTo != "" {
			stateVal = state + "|" + returnTo
		}
		c.Cookie(&fiber.Cookie{
			Name:     "eop_oauth_state",
			Value:    stateVal,
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   s.SecureCookies,
			MaxAge:   600,
		})
		return c.Redirect(p.AuthCodeURL(state))
	})

	g.Get("/"+name+"/callback", func(c *fiber.Ctx) error {
		got := c.Query("state")
		stored := c.Cookies("eop_oauth_state")

		nonce := stored
		returnTo := ""
		if i := strings.Index(stored, "|"); i > 0 {
			nonce = stored[:i]
			returnTo = stored[i+1:]
		}
		if got == "" || got != nonce {
			return httperr.BadRequest(c, "state_mismatch", "state mismatch")
		}
		code := c.Query("code")
		if code == "" {

			if c.Query("error") == "access_denied" {
				return httperr.BadRequest(c, "user_denied", "user denied access")
			}
			return httperr.BadRequest(c, "missing_code", "missing code")
		}
		user, err := p.Exchange(c.Context(), code)
		if err != nil {
			s.Logger.Warn("oauth exchange failed", zap.String("provider", name), zap.Error(err))
			return httperr.BadGateway(c, "oauth_exchange_failed", "oauth exchange failed")
		}
		if user.Email == "" {
			return httperr.BadRequest(c, "email_not_verified", "provider did not return verified email")
		}

		userUUID, linkErr := newOAuthAppService(s).UpsertOAuthUser(c.Context(), name, oauthapp.ExternalUser{
			Subject: user.Subject, Email: user.Email, Name: user.Name, Login: user.Login,
		})
		if linkErr != nil {
			s.Logger.Warn("oauth link failed", zap.String("provider", name), zap.Error(linkErr))
			if errors.Is(linkErr, oauthapp.ErrIdentityLinkRequired) {

				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error":    "identity_link_required",
					"detail":   "an account exists with this email; sign in with password to link",
					"email":    user.Email,
					"provider": name,
				})
			}
			return httperr.Conflict(c, "oauth_link_failed", "could not link account")
		}

		tv, _ := TokenVersion(c.Context(), s.Pool, userUUID)
		tok, err := IssueJWT(s.JWTSecret, userUUID.String(), user.Email, name, tv, tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}
		setHandoffCookie(c, s, tok)

		c.Cookie(&fiber.Cookie{Name: "eop_oauth_state", Value: "", MaxAge: -1, HTTPOnly: true, Secure: s.SecureCookies})

		dest := strings.TrimRight(s.PublicURL, "/") + "/auth/complete"
		if returnTo != "" {
			dest += "?return_to=" + base64URLEncode(returnTo)
		}
		return c.Redirect(dest, fiber.StatusFound)
	})
}

func setHandoffCookie(c *fiber.Ctx, s Service, tok string) {
	c.Cookie(&fiber.Cookie{
		Name:     handoffCookieName,
		Value:    tok,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   s.SecureCookies,
		MaxAge:   int(handoffTTL.Seconds()),
	})
}

func handleSessionHandoff(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tok := c.Cookies(handoffCookieName)
		if tok == "" {
			return httperr.NotFound(c, "no_handoff", "no pending session handoff")
		}

		c.Cookie(&fiber.Cookie{
			Name:     handoffCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   0,
			Expires:  time.Unix(0, 0).UTC(),
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   s.SecureCookies,
		})
		claims, err := ParseJWT(s.JWTSecret, tok)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_handoff", "invalid handoff token")
		}

		if claims.IssuedAt != nil {
			if time.Since(claims.IssuedAt.Time) > handoffTTL+5*time.Second {
				return httperr.Unauthorized(c, "expired_handoff", "handoff token too old")
			}
		}
		return c.JSON(fiber.Map{"token": tok})
	}
}

func RegisterIdentitiesRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/me", Middleware(s.JWTSecret, s.Pool))
	g.Get("/identities", handleListIdentities(s))
	g.Delete("/identities/:id", handleDeleteIdentity(s))

	if s.WebAuthn != nil {
		g.Get("/passkeys", handleListPasskeys(s))
		g.Delete("/passkeys/:id", handleDeletePasskey(s))
	}
}

type IdentityRow struct {
	ID         uuid.UUID  `json:"id"`
	Provider   string     `json:"provider"`
	Subject    string     `json:"subject"`
	Email      string     `json:"email,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func handleListIdentities(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		if s.Pool == nil {
			return c.JSON(fiber.Map{"identities": []IdentityRow{}})
		}
		rows, err := s.Pool.Query(c.Context(), `
			SELECT id, provider, subject, COALESCE(email, ''), created_at
			FROM user_identities
			WHERE user_id = $1
			ORDER BY created_at DESC
		`, uid)
		if err != nil {
			return httperr.Internal(c)
		}
		defer rows.Close()
		out := []IdentityRow{}
		for rows.Next() {
			var r IdentityRow
			if err := rows.Scan(&r.ID, &r.Provider, &r.Subject, &r.Email, &r.CreatedAt); err != nil {
				return httperr.Internal(c)
			}
			out = append(out, r)
		}
		return c.JSON(fiber.Map{"identities": out})
	}
}

func handleDeleteIdentity(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		idStr := c.Params("id")
		identityID, err := uuid.Parse(idStr)
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid identity id")
		}
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_required", "identities require database")
		}

		factors, err := CountAuthFactors(c.Context(), s.Pool, uid, &identityID, nil)
		if err != nil {
			return s.internalErr(c, err)
		}
		if factors.Total() == 0 {

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "last_auth_factor",
				"detail": "cannot remove last sign-in method; set a password or add a passkey first",
			})
		}

		res, err := s.Pool.Exec(c.Context(),
			`DELETE FROM user_identities WHERE id = $1 AND user_id = $2`,
			identityID, uid,
		)
		if err != nil {
			return s.internalErr(c, err)
		}
		if res.RowsAffected() == 0 {
			return httperr.NotFound(c, "not_found", "identity not found")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func handleListPasskeys(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		rows, err := s.WebAuthn.ListPasskeys(c.Context(), uid)
		if err != nil {
			return s.internalErr(c, err)
		}
		return c.JSON(fiber.Map{"passkeys": rows})
	}
}

func handleDeletePasskey(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		idStr := c.Params("id")
		passkeyID, err := uuid.Parse(idStr)
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid passkey id")
		}

		credID, err := s.WebAuthn.PasskeyCredentialIDForUser(c.Context(), uid, passkeyID)
		if err != nil {
			if errors.Is(err, ErrPasskeyNotFound) {
				return httperr.NotFound(c, "not_found", "passkey not found")
			}
			return s.internalErr(c, err)
		}

		factors, err := CountAuthFactors(c.Context(), s.Pool, uid, nil, credID)
		if err != nil {
			return s.internalErr(c, err)
		}
		if factors.Total() == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "last_auth_factor",
				"detail": "cannot remove last sign-in method; set a password or add an identity first",
			})
		}

		if err := s.WebAuthn.DeletePasskey(c.Context(), uid, passkeyID); err != nil {
			if errors.Is(err, ErrPasskeyNotFound) {
				return httperr.NotFound(c, "not_found", "passkey not found")
			}
			return s.internalErr(c, err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func handleWebAuthnRegisterBegin(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		wa := newWebAuthnApp(s.WebAuthn)
		creation, sid, err := wa.RegisterBegin(c.Context(), uid)
		if err != nil {
			s.Logger.Warn("webauthn begin register failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_begin_failed", err.Error())
		}
		return c.JSON(fiber.Map{
			"options":    creation,
			"session_id": sid,
		})
	}
}

type webauthnFinishRegisterReq struct {
	SessionID   string          `json:"session_id"`
	Attestation json.RawMessage `json:"attestation"`
	Nickname    string          `json:"nickname"`
}

func handleWebAuthnRegisterFinish(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		var req webauthnFinishRegisterReq
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if req.SessionID == "" || len(req.Attestation) == 0 {
			return httperr.BadRequest(c, "missing_fields", "session_id and attestation required")
		}
		if err := newWebAuthnApp(s.WebAuthn).RegisterFinish(c.Context(), uid, req.SessionID, req.Attestation, req.Nickname); err != nil {
			s.Logger.Warn("webauthn finish register failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_finish_failed", err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

type webauthnLoginBeginReq struct {
	Email *string `json:"email"`
}

func handleWebAuthnLoginBegin(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req webauthnLoginBeginReq

		_ = c.BodyParser(&req)
		assertion, sid, err := newWebAuthnApp(s.WebAuthn).LoginBegin(c.Context(), req.Email)
		if err != nil {
			s.Logger.Warn("webauthn begin login failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_begin_failed", err.Error())
		}
		return c.JSON(fiber.Map{
			"options":    assertion,
			"session_id": sid,
		})
	}
}

type webauthnLoginFinishReq struct {
	SessionID string          `json:"session_id"`
	Assertion json.RawMessage `json:"assertion"`

	ReturnTo string `json:"return_to"`
}

func handleWebAuthnLoginFinish(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req webauthnLoginFinishReq
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if req.SessionID == "" || len(req.Assertion) == 0 {
			return httperr.BadRequest(c, "missing_fields", "session_id and assertion required")
		}
		uid, err := newWebAuthnApp(s.WebAuthn).LoginFinish(c.Context(), req.SessionID, req.Assertion)
		if err != nil {
			s.Logger.Warn("webauthn finish login failed", zap.Error(err))
			return httperr.Unauthorized(c, "webauthn_finish_failed", err.Error())
		}

		var email string
		if s.Pool != nil {
			_ = s.Pool.QueryRow(c.Context(),
				`SELECT COALESCE(email, '') FROM users WHERE id = $1`, uid,
			).Scan(&email)
		}
		tv, _ := TokenVersion(c.Context(), s.Pool, uid)
		tok, err := IssueJWT(s.JWTSecret, uid.String(), email, "passkey", tv, tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}

		setHandoffCookie(c, s, tok)
		return c.JSON(fiber.Map{"token": tok, "user_id": uid.String()})
	}
}

func (s Service) internalErr(c *fiber.Ctx, err error) error {
	rid, _ := c.Locals("requestid").(string)
	s.Logger.Error("internal error",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()),
		zap.String("rid", rid),
		zap.Error(err),
	)
	return httperr.Internal(c)
}

func registerDevToken(g fiber.Router, s Service) {
	g.Post("/dev-token", func(c *fiber.Ctx) error {
		userIDStr := c.Query("user_id")
		var userID uuid.UUID
		if userIDStr == "" {
			userID = uuid.New()
		} else {
			parsed, err := uuid.Parse(userIDStr)
			if err != nil {
				return httperr.BadRequest(c, "invalid_user_id", "user_id must be uuid")
			}
			userID = parsed
		}
		email := fmt.Sprintf("dev-%s@local.eop", userID.String()[:8])
		if err := s.Users.Upsert(c.Context(), userID, email, ""); err != nil {
			s.Logger.Warn("user upsert failed (continuing with token)", zap.Error(err))
		}
		tv, _ := TokenVersion(c.Context(), s.Pool, userID)
		tok, err := IssueJWT(s.JWTSecret, userID.String(), email, "dev", tv, tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"token": tok, "user_id": userID.String()})
	})
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func base64URLEncode(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
