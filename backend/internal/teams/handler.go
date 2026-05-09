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

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/mailer"
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

// --- Validators ---

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

// --- Service ---

type Service struct {
	Pool          *pgxpool.Pool
	JWTSecret     string
	Logger        *zap.Logger
	InviteOnly    bool // регистрация только по invite (первый user всегда может — bootstrap)
	BetaTeamLimit int  // 0 = без лимита, иначе — максимум команд для бета-программы
	Mailer        mailer.Mailer
	PublicURL     string // base URL дашборда для invite-ссылок в письме
}

// internalErr — единая точка для 500-ответов. Логируем полный текст,
// клиенту отдаём generic message + request_id для корреляции.
func (s Service) internalErr(c *fiber.Ctx, err error) error {
	rid, _ := c.Locals("requestid").(string)
	s.Logger.Error("internal error",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()),
		zap.String("rid", rid),
		zap.Error(err),
	)
	return c.Status(500).JSON(fiber.Map{"error": "internal error", "request_id": rid})
}

// --- Routes ---

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
	g.Delete("/admin/teams/:id", s.handleAdminDeleteTeam)
	g.Delete("/admin/users/:id", s.handleAdminDeleteUser)
	g.Patch("/admin/users/:id", s.handleAdminUpdateUser)
	g.Post("/admin/teams/:id/members", s.handleAdminAddMember)
	g.Patch("/admin/teams/:id/subscription", s.handleSetSubscription)
	g.Get("/admin/teams/:id/payments", s.handleListPayments)
}

// --- Helpers (используются handler'ами всех доменных файлов) ---

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
