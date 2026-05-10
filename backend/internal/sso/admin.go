package sso

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

// AdminService — bundle для admin SSO endpoints. teamRoleCheck inject'ится
// чтобы избежать import cycle (teams imports sso для admin routes ниже).
type AdminService struct {
	Pool     *pgxpool.Pool
	Registry *Registry
	Logger   *zap.Logger
}

// RegisterAdminRoutes — owner/admin endpoints для SSO config CRUD.
// Mount'ится под /v1/teams/:id/sso, защищается auth middleware на caller
// side. Per-endpoint role check внутри.
func (s AdminService) RegisterAdminRoutes(g fiber.Router) {
	g.Get("/", s.handleGet)
	g.Put("/", s.handleSave)
	g.Delete("/", s.handleDelete)
	g.Post("/test", s.handleTestDiscovery)
}

// configResponse — конфиг без секретов. client_secret никогда не возвращается
// в GET — это write-only.
type configResponse struct {
	TeamID         uuid.UUID `json:"team_id"`
	Provider       string    `json:"provider"`
	Enabled        bool      `json:"enabled"`
	OIDCIssuer     string    `json:"oidc_issuer,omitempty"`
	OIDCClientID   string    `json:"oidc_client_id,omitempty"`
	HasClientSecret bool     `json:"has_client_secret"` // только индикатор
	OIDCScopes     []string  `json:"oidc_scopes,omitempty"`
	AllowedDomains []string  `json:"allowed_domains,omitempty"`
	JITProvision   bool      `json:"jit_provision"`
	JITRole        string    `json:"jit_role"`
}

func toResponse(c *Config) configResponse {
	return configResponse{
		TeamID:          c.TeamID,
		Provider:        string(c.Provider),
		Enabled:         c.Enabled,
		OIDCIssuer:      c.OIDCIssuer,
		OIDCClientID:    c.OIDCClientID,
		HasClientSecret: c.OIDCClientSecret != "",
		OIDCScopes:      c.OIDCScopes,
		AllowedDomains:  c.AllowedDomains,
		JITProvision:    c.JITProvision,
		JITRole:         c.JITRole,
	}
}

// requireOwnerOrAdmin — проверяет что юзер owner|admin в team. Возвращает
// httperr response если нет, или nil если ok.
func requireOwnerOrAdmin(c *fiber.Ctx, pool *pgxpool.Pool, teamID uuid.UUID) error {
	claims := auth.ClaimsFromCtx(c)
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
	}
	var role string
	err = pool.QueryRow(c.Context(),
		`SELECT role FROM team_members WHERE user_id=$1 AND team_id=$2`,
		uid, teamID).Scan(&role)
	if err != nil {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	if role != "owner" && role != "admin" {
		return httperr.Forbidden(c, "role_insufficient", "owner/admin required")
	}
	return nil
}

func (s AdminService) handleGet(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	if err := requireOwnerOrAdmin(c, s.Pool, teamID); err != nil {
		return err
	}
	cfg, err := LoadConfig(c.Context(), s.Pool, teamID)
	if errors.Is(err, ErrConfigNotFound) {
		return c.JSON(fiber.Map{"configured": false})
	}
	if err != nil {
		return httperr.Internal(c)
	}
	return c.JSON(fiber.Map{"configured": true, "config": toResponse(cfg)})
}

type saveRequest struct {
	Provider         string   `json:"provider"`           // "oidc"
	Enabled          bool     `json:"enabled"`
	OIDCIssuer       string   `json:"oidc_issuer"`
	OIDCClientID     string   `json:"oidc_client_id"`
	OIDCClientSecret string   `json:"oidc_client_secret"` // optional, если пусто — keep existing
	OIDCScopes       []string `json:"oidc_scopes"`
	AllowedDomains   []string `json:"allowed_domains"`
	JITProvision     *bool    `json:"jit_provision"`
	JITRole          string   `json:"jit_role"`
}

func (s AdminService) handleSave(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	if err := requireOwnerOrAdmin(c, s.Pool, teamID); err != nil {
		return err
	}
	var req saveRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}

	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		provider = "oidc"
	}
	if provider != "oidc" {
		return httperr.BadRequest(c, "unsupported_provider",
			"only 'oidc' supported in this release (SAML coming)")
	}
	if req.OIDCIssuer == "" || req.OIDCClientID == "" {
		return httperr.BadRequest(c, "incomplete_oidc",
			"oidc_issuer and oidc_client_id required")
	}

	cfg, err := LoadConfig(c.Context(), s.Pool, teamID)
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		return httperr.Internal(c)
	}
	if cfg == nil {
		cfg = &Config{TeamID: teamID, Provider: ProviderOIDC, JITProvision: true, JITRole: "member"}
	}

	cfg.Provider = ProviderOIDC
	cfg.Enabled = req.Enabled
	cfg.OIDCIssuer = strings.TrimRight(req.OIDCIssuer, "/")
	cfg.OIDCClientID = req.OIDCClientID
	if req.OIDCClientSecret != "" {
		// Empty в request = keep existing (для UI: чтобы owner не вводил secret
		// каждый раз при редактировании other fields).
		cfg.OIDCClientSecret = req.OIDCClientSecret
	}
	if cfg.OIDCClientSecret == "" {
		return httperr.BadRequest(c, "missing_client_secret",
			"client_secret required on first save")
	}
	cfg.OIDCScopes = normalizeScopes(req.OIDCScopes)
	cfg.AllowedDomains = normalizeDomains(req.AllowedDomains)
	if req.JITProvision != nil {
		cfg.JITProvision = *req.JITProvision
	}
	if req.JITRole != "" {
		role := strings.ToLower(strings.TrimSpace(req.JITRole))
		if role != "owner" && role != "admin" && role != "member" {
			return httperr.BadRequest(c, "invalid_jit_role",
				"jit_role must be owner|admin|member")
		}
		cfg.JITRole = role
	}

	if err := SaveConfig(c.Context(), s.Pool, cfg); err != nil {
		s.Logger.Error("sso config save failed", zap.Error(err))
		return httperr.Internal(c)
	}
	s.Registry.Invalidate(teamID)

	return c.JSON(fiber.Map{"config": toResponse(cfg)})
}

func (s AdminService) handleDelete(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	if err := requireOwnerOrAdmin(c, s.Pool, teamID); err != nil {
		return err
	}
	if err := DeleteConfig(c.Context(), s.Pool, teamID); err != nil {
		return httperr.Internal(c)
	}
	s.Registry.Invalidate(teamID)
	return c.JSON(fiber.Map{"deleted": true})
}

// handleTestDiscovery — проверяет, что issuer URL валиден (well-known
// endpoint достижим, ID token signing keys publish'ятся). Не делает actual
// flow, нужен только для UI "Test connection" кнопки.
func (s AdminService) handleTestDiscovery(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	if err := requireOwnerOrAdmin(c, s.Pool, teamID); err != nil {
		return err
	}
	cfg, err := LoadConfig(c.Context(), s.Pool, teamID)
	if errors.Is(err, ErrConfigNotFound) {
		return httperr.NotFound(c, "sso_not_configured", "SSO not configured")
	}
	if err != nil {
		return httperr.Internal(c)
	}
	if _, err := NewOIDCProvider(c.Context(), cfg, ""); err != nil {
		return httperr.BadGateway(c, "discovery_failed", err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "issuer": cfg.OIDCIssuer})
}

func normalizeScopes(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func normalizeDomains(in []string) []string {
	out := make([]string, 0, len(in))
	for _, d := range in {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" {
			out = append(out, d)
		}
	}
	return out
}
