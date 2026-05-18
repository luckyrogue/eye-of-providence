package teams

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	InviteOnly    bool
	BetaTeamLimit int
	Mailer        mailer.Mailer
	PublicURL     string
	Webhooks      WebhookDispatcher
	Plans         plans.Service
	Audit         audit.Service

	AuthProviders []string

	PasskeyEnabled bool

	TemplateStore *mailer.PGTemplateStore

	EventStore EventStoreLike
}

func (s Service) actorInfo(c *fiber.Ctx) (uuid.UUID, string) {
	uid := userID(c)
	var email string
	if s.Pool != nil {
		_ = s.Pool.QueryRow(c.Context(), "SELECT email FROM users WHERE id=$1", uid).Scan(&email)
	}
	return uid, email
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

func RegisterRoutes(app *fiber.App, s Service) {

	a := app.Group("/v1/auth")
	a.Post("/register", s.handleRegister)
	a.Post("/login", s.handleLogin)
	a.Get("/config", s.handleAuthConfig)
	app.Get("/v1/invites/:code", s.handleInvitePreview)

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
	g.Post("/commits", s.handleIngestCommit)

	g.Post("/invites/:code/accept", s.handleInviteAccept)

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
	return s.membersApp().TeamRole(ctx, userID, teamID)
}

func (s Service) isSuperAdmin(c *fiber.Ctx) bool {
	uid := userID(c)
	var r string
	err := s.Pool.QueryRow(c.Context(), "SELECT global_role FROM users WHERE id=$1", uid).Scan(&r)
	return err == nil && r == "super_admin"
}

