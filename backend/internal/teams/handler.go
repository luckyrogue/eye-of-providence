// Package teams — auth (email+password), команды, invites, projects, commits.
//
// Структура файлов (по доменам):
//
//	handler.go      — Service, RegisterRoutes, validators, helpers
//	auth.go         — register / login / authConfig
//	password_reset.go — forgot / reset password
//	teams.go        — list / create / detail / update / delete / beta info
//	members.go      — list / summary / role-update / remove
//	invites.go      — link + email invites lifecycle
//	projects.go     — list / create projects
//	commits.go      — git commit ingest + queries
//	admin.go        — super_admin endpoints
//
// Роли:
//
//	global: users.global_role ∈ {super_admin, user}
//	per-team: team_members.role ∈ {owner, admin, member}
package teams

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/mailer"
	"github.com/eye-of-providence/backend/internal/plans"
)

const (
	minPasswordLen    = 8
	maxPasswordLen    = 256
	maxDisplayNameLen = 64
	maxTeamNameLen    = 100
	maxProjectNameLen = 200

	// teamCreationLockID — sentinel-id для pg_advisory_xact_lock, защищающий
	// инвариант "1 owner = 1 team" + лимит beta-команд от race'ов.
	// Произвольное число, фиксированное на весь жизненный цикл проекта.
	teamCreationLockID int64 = 8331_2026_001

	tokenTTL = 14 * 24 * time.Hour
)

func validateEmail(s string) (string, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || len(s) > 254 {
		return "", false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return "", false
	}
	return addr.Address, true
}

func validatePassword(s string) bool {
	return len(s) >= minPasswordLen && len(s) <= maxPasswordLen
}

func validateDisplayName(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if len(s) > maxDisplayNameLen {
		return "", false
	}
	if strings.ContainsAny(s, "\r\n\t") {
		return "", false
	}
	return s, true
}

type Service struct {
	Pool          *pgxpool.Pool
	JWTSecret     string
	Logger        *zap.Logger
	InviteOnly    bool // регистрация только по invite (первый user всегда может — bootstrap)
	BetaTeamLimit int  // 0 = без лимита, иначе — максимум команд для бета-программы
	Mailer        mailer.Mailer
	PublicURL     string            // base URL дашборда для invite-ссылок в письме
	Webhooks      WebhookDispatcher // nil — webhook delivery выключена (in-memory mode)
	Plans         plans.Service     // feature-gate: max users/team при Enforce=true
	Audit         audit.Service     // append-only лог критичных действий
	// AuthProviders — массив включённых OAuth-провайдеров ("github", "google", ...).
	// Конструируется в main.go на основе env. Отдаётся фронту через GET /v1/auth/config
	// для conditional рендеринга кнопок.
	AuthProviders []string
	// PasskeyEnabled — true если WebAuthn настроен (RPID + Origins). Frontend
	// рендерит passkey-кнопку только если true.
	PasskeyEnabled bool
	// TemplateStore — Postgres-backed override store для transactional email
	// templates (Phase 3 admin). nil = admin endpoints возвращают 503 и
	// Mailer.Send продолжает использовать только embedded baseline.
	TemplateStore *mailer.PGTemplateStore
}

// actorInfo — берёт текущего user_id + email для audit-trail. Email
// денормализован чтобы audit-row сохранялся даже если user удалён.
func (s Service) actorInfo(c *fiber.Ctx) (uuid.UUID, string) {
	uid := userID(c)
	var email string
	if s.Pool != nil {
		_ = s.Pool.QueryRow(c.Context(), "SELECT email FROM users WHERE id=$1", uid).Scan(&email)
	}
	return uid, email
}

// internalErr — единая точка для 500-ответов. Логируем полный текст,
// клиенту отдаём generic message + request_id для корреляции (через httperr).
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

func RegisterRoutes(app *fiber.App, s Service) {
	// Public
	a := app.Group("/v1/auth")
	a.Post("/register", s.handleRegister)
	a.Post("/login", s.handleLogin)
	a.Post("/forgot-password", s.handleForgotPassword)
	a.Post("/reset-password", s.handleResetPassword)
	a.Get("/config", s.handleAuthConfig)
	app.Get("/v1/invites/:code", s.handleInvitePreview)

	// Authed
	g := app.Group("/v1", auth.Middleware(s.JWTSecret, s.Pool))

	g.Get("/teams", s.handleListMyTeams)
	g.Post("/teams", s.handleCreateTeam)
	g.Get("/beta/info", s.handleBetaInfo)

	g.Patch("/me/email", s.handleChangeMyEmail)
	g.Patch("/me/password", s.handleChangeMyPassword)
	g.Patch("/me/name", s.handleChangeMyName)

	t := g.Group("/teams/:id")
	t.Get("/", s.handleTeamDetail)
	t.Patch("/", s.handleUpdateTeam)
	t.Delete("/", s.handleDeleteTeam)
	t.Get("/members", s.handleListMembers)
	t.Patch("/members/:user_id", s.handleUpdateMemberRole)
	t.Delete("/members/:user_id", s.handleRemoveMember)
	t.Post("/invites", s.handleCreateInvite)
	t.Get("/summary", s.handleTeamSummary)

	t.Get("/projects", s.handleListProjects)
	t.Post("/projects", s.handleCreateProject)
	t.Get("/projects/:project_id/commits", s.handleProjectCommits)

	t.Get("/commits", s.handleTeamCommits)
	g.Post("/commits", s.handleIngestCommit) // user pushes commit info через CLI/git hook

	g.Post("/invites/:code/accept", s.handleInviteAccept)

	// Super admin — full management. Все защищены requireSuperAdmin внутри.
	g.Get("/admin/teams", s.handleAdminListAllTeams)
	g.Get("/admin/users", s.handleAdminListAllUsers)
	g.Get("/admin/stats", s.handleAdminStats)
	g.Get("/admin/revenue", s.handleAdminRevenue)
	g.Get("/admin/sso", s.handleAdminSSOList)
	g.Post("/admin/sso/:id/disable", s.handleAdminSSODisable)
	g.Get("/admin/audit", s.handleAdminAudit)
	g.Delete("/admin/teams/:id", s.handleAdminDeleteTeam)
	g.Delete("/admin/users/:id", s.handleAdminDeleteUser)
	g.Patch("/admin/users/:id", s.handleAdminUpdateUser)
	g.Post("/admin/teams/:id/members", s.handleAdminAddMember)
	g.Patch("/admin/teams/:id/subscription", s.handleSetSubscription)
	g.Get("/admin/teams/:id/payments", s.handleListPayments)

	// Phase 3 admin: email templates, team flags, plan-limit overrides,
	// cross-team webhooks/api-tokens. Все защищены requireSuperAdmin внутри.
	g.Get("/admin/email-templates", s.handleAdminListEmailTemplates)
	g.Get("/admin/email-templates/:key/:locale", s.handleAdminGetEmailTemplate)
	g.Put("/admin/email-templates/:key/:locale", s.handleAdminUpsertEmailTemplate)
	g.Delete("/admin/email-templates/:key/:locale", s.handleAdminDeleteEmailTemplate)

	g.Get("/admin/teams/:id/flags", s.handleAdminGetTeamFlags)
	g.Patch("/admin/teams/:id/flags", s.handleAdminPatchTeamFlags)
	g.Get("/admin/teams/:id/plan-limits", s.handleAdminGetTeamPlanLimits)
	g.Patch("/admin/teams/:id/plan-limits", s.handleAdminPatchTeamPlanLimits)

	g.Get("/admin/webhooks", s.handleAdminListAllWebhooks)
	g.Get("/admin/api-tokens", s.handleAdminListAllAPITokens)
}

func userID(c *fiber.Ctx) uuid.UUID {
	claims := auth.ClaimsFromCtx(c)
	uid, _ := uuid.Parse(claims.UserID)
	return uid
}

func (s Service) teamRole(ctx context.Context, userID, teamID uuid.UUID) (string, bool) {
	var role string
	err := s.Pool.QueryRow(ctx,
		"SELECT role FROM team_members WHERE user_id=$1 AND team_id=$2",
		userID, teamID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return role, true
}

func (s Service) addMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO NOTHING`, teamID, userID, role)
	return err
}

func (s Service) isSuperAdmin(c *fiber.Ctx) bool {
	uid := userID(c)
	var r string
	err := s.Pool.QueryRow(c.Context(), "SELECT global_role FROM users WHERE id=$1", uid).Scan(&r)
	return err == nil && r == "super_admin"
}

func randomCode(bytes int) string {
	b := make([]byte, bytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
