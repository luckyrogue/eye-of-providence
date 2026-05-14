package sso

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

const tokenTTL = 14 * 24 * time.Hour

// Service — DI bundle для SSO handler'ов.
type Service struct {
	Pool      *pgxpool.Pool
	Registry  *Registry
	Logger    *zap.Logger
	JWTSecret string
	PublicURL string // dashboard URL для default redirect_uri
}

// RegisterRoutes — публичные SSO endpoints (без auth middleware — это
// вход в систему). Admin endpoints для config CRUD регистрируются
// отдельно в teams/admin (требуют JWT+owner role).
func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/sso")
	g.Post("/start", s.handleStart)                  // POST {team_id, return_to} → 302 IdP authorize URL
	g.Get("/oidc/callback", s.handleOIDCCallback)    // IdP redirect_uri
}

// handleStart — SP-initiated start flow. Body:
//
//	{"team_id": "uuid", "return_to": "/dashboard"}
//
// Возвращает {"authorize_url": "..."} — frontend делает window.location =
// authorize_url. Можно сделать прямой 302, но JSON удобнее для SPA: дашборд
// может вставить trackerы / loading состояние перед redirect'ом.
func (s Service) handleStart(c *fiber.Ctx) error {
	var req struct {
		TeamID   string `json:"team_id"`
		ReturnTo string `json:"return_to"`
	}
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "team_id must be uuid")
	}

		url, err := newSSOStartService(s).AuthorizeURL(c.Context(), teamID, req.ReturnTo)
		if errors.Is(err, ErrConfigNotFound) {
			return httperr.NotFound(c, "sso_not_configured", "SSO not configured for this team")
		}
		if err != nil {
			s.Logger.Error("sso start failed",
				zap.String("team", teamID.String()), zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"authorize_url": url})
}

// handleOIDCCallback — IdP redirect endpoint. Query: ?code=...&state=...
// Если error param — IdP отклонил (e.g. user denied consent).
//
// Success → redirect на `{public_url}/sso-callback#token=...&return_to=...`.
// Fragment чтобы JWT не светился в server logs / referrer headers; frontend
// ловит на client-side route, кладёт в localStorage, делает navigate(returnTo).
func (s Service) handleOIDCCallback(c *fiber.Ctx) error {
	if e := c.Query("error"); e != "" {
		// IdP вернул ошибку (user denied / config issue).
		s.Logger.Warn("idp returned error",
			zap.String("error", e),
			zap.String("description", c.Query("error_description")))
		return httperr.BadGateway(c, "idp_error", e)
	}
	code := c.Query("code")
	stateValue := c.Query("state")
	if code == "" || stateValue == "" {
		return httperr.BadRequest(c, "missing_params", "code and state required")
	}

	state, err := ConsumeState(c.Context(), s.Pool, stateValue)
	if err != nil {
		return httperr.BadRequest(c, "state_invalid", "invalid or expired state")
	}

	prov, err := s.Registry.Get(c.Context(), state.TeamID)
	if err != nil {
		s.Logger.Error("sso registry on callback failed", zap.Error(err))
		return httperr.Internal(c)
	}

	ident, err := prov.Exchange(c.Context(), code, state.Nonce)
	if err != nil {
		s.Logger.Warn("oidc exchange failed",
			zap.String("team", state.TeamID.String()), zap.Error(err))
		return httperr.BadGateway(c, "oidc_exchange_failed", "OIDC code exchange failed")
	}
	if err := prov.CheckEmailDomain(ident.Email); err != nil {
		s.Logger.Warn("email domain rejected",
			zap.String("email", ident.Email), zap.Error(err))
		return httperr.Forbidden(c, "domain_not_allowed", err.Error())
	}
	if !ident.EmailVerified {
		// Конфигуратоp IdP'а должен email verify'ить; иначе attacker может
		// подделать unverified email со совпадающим домен'ом.
		return httperr.Forbidden(c, "email_unverified", "IdP marked email as unverified")
	}

	cfg, err := LoadConfig(c.Context(), s.Pool, state.TeamID)
	if err != nil {
		return httperr.Internal(c)
	}

	user, isNew, err := ProvisionUser(c.Context(), s.Pool, state.TeamID, cfg, ident)
	if errors.Is(err, ErrJITDisabled) {
		return httperr.Forbidden(c, "jit_disabled",
			"user not provisioned — ask team admin to invite you first")
	}
	if err != nil {
		s.Logger.Error("sso provision user failed", zap.Error(err))
		return httperr.Internal(c)
	}

	tok, err := auth.IssueJWT(s.JWTSecret, user.ID.String(), user.Email,
		"sso", user.TokenVersion, tokenTTL)
	if err != nil {
		s.Logger.Error("sso issue jwt failed", zap.Error(err))
		return httperr.Internal(c)
	}

	// Touch best-effort (non-fatal).
	go Touch(c.UserContext(), s.Pool, user.ID, ident)

	s.Logger.Info("sso login success",
		zap.String("user", user.ID.String()),
		zap.String("team", state.TeamID.String()),
		zap.Bool("jit_created", isNew))

	// Redirect через #fragment.
	rt := state.ReturnTo
	if rt == "" {
		rt = "/dashboard"
	}
	redirect := s.PublicURL + "/sso-callback#" +
		"token=" + tok +
		"&user_id=" + user.ID.String() +
		"&return_to=" + escapePath(rt)
	return c.Redirect(redirect)
}

// escapePath — минимальный URL-escape для return_to. Не используем
// url.QueryEscape потому что нам нужен path-friendly формат (#fragment
// не должен ломать parse).
func escapePath(s string) string {
	// Только опасные символы. Leading "/" allowed.
	r := strings.NewReplacer(
		"#", "%23",
		"&", "%26",
		" ", "%20",
	)
	return r.Replace(s)
}
