// Package teams — auth (email+password), команды, invites, projects, commits.
//
// Роли:
//   global: users.global_role ∈ {super_admin, user}
//   per-team: team_members.role ∈ {owner, admin, member}
package teams

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/store"
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

const tokenTTL = 14 * 24 * time.Hour

type Service struct {
	Pool          *pgxpool.Pool
	JWTSecret     string
	Logger        *zap.Logger
	InviteOnly    bool // регистрация только по invite (первый user всегда может — bootstrap)
	BetaTeamLimit int  // 0 = без лимита, иначе — максимум команд для бета-программы
}

func RegisterRoutes(app *fiber.App, s Service) {
	// Public
	a := app.Group("/v1/auth")
	a.Post("/register", s.handleRegister)
	a.Post("/login", s.handleLogin)
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

// --- Public auth config (для UI) ---

func (s Service) handleAuthConfig(c *fiber.Ctx) error {
	var userCount int
	if s.Pool != nil {
		_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM users").Scan(&userCount)
	}
	return c.JSON(fiber.Map{
		"invite_only":   s.InviteOnly,
		"is_first_user": userCount == 0,
	})
}

// --- Auth: register / login ---

type registerReq struct {
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
	Password    string  `json:"password"`
	InviteCode  *string `json:"invite_code,omitempty"`
}

func (s Service) handleRegister(c *fiber.Ctx) error {
	if s.Pool == nil {
		return c.Status(503).JSON(fiber.Map{"error": "auth requires postgres"})
	}
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	email, ok := validateEmail(req.Email)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "valid email required"})
	}
	req.Email = email
	if !validatePassword(req.Password) {
		return c.Status(400).JSON(fiber.Map{"error": "password 8..256 chars"})
	}
	if req.DisplayName == "" {
		req.DisplayName = req.Email
	}
	dn, ok := validateDisplayName(req.DisplayName)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "display_name 1..64 chars, no newlines"})
	}
	req.DisplayName = dn

	// Подсчёт пользователей в системе.
	// Первый user может зарегаться без invite (bootstrap), он же станет super_admin.
	var userCount int
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM users").Scan(&userCount)
	isFirstUser := userCount == 0

	if s.InviteOnly && !isFirstUser {
		if req.InviteCode == nil || *req.InviteCode == "" {
			return c.Status(403).JSON(fiber.Map{"error": "регистрация только по приглашению"})
		}
		// Проверяем валидность ДО создания юзера
		if _, err := s.findInvite(c.Context(), *req.InviteCode); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "приглашение невалидно или устарело"})
		}
	}

	user, err := auth.CreateUser(c.Context(), s.Pool, req.Email, req.DisplayName, req.Password, nil)
	if err != nil {
		return c.Status(409).JSON(fiber.Map{"error": "email уже занят (или ошибка БД)"})
	}

	// Первый user становится super_admin
	if isFirstUser {
		_, _ = s.Pool.Exec(c.Context(),
			"UPDATE users SET global_role='super_admin' WHERE id=$1", user.ID)
		s.Logger.Info("first user promoted to super_admin", zap.String("email", user.Email))
	}

	// Атомарно: invite consumes (защищает от race), затем добавляем юзера.
	var joinedTeam *uuid.UUID
	if req.InviteCode != nil && *req.InviteCode != "" {
		teamID, err := s.consumeInvite(c.Context(), *req.InviteCode, user.ID)
		if err == nil {
			_ = s.addMember(c.Context(), teamID, user.ID, "member")
			joinedTeam = &teamID
		}
	}

	tv, _ := auth.TokenVersion(c.Context(), s.Pool, user.ID)
	tok, err := auth.IssueJWT(s.JWTSecret, user.ID.String(), user.Email, "password", tv, tokenTTL)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"token":        tok,
		"user_id":      user.ID,
		"display_name": user.DisplayName,
		"team_id":      joinedTeam,
	})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s Service) handleLogin(c *fiber.Ctx) error {
	if s.Pool == nil {
		return c.Status(503).JSON(fiber.Map{"error": "auth requires postgres"})
	}
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	email, ok := validateEmail(req.Email)
	if !ok || !validatePassword(req.Password) {
		// Не подсказываем, что именно не так — это login surface.
		return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
	}
	req.Email = email
	user, err := auth.FindUserByEmail(c.Context(), s.Pool, req.Email)
	if err != nil || !auth.VerifyPassword(user.PasswordHash, req.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
	}
	tv, _ := auth.TokenVersion(c.Context(), s.Pool, user.ID)
	tok, err := auth.IssueJWT(s.JWTSecret, user.ID.String(), user.Email, "password", tv, tokenTTL)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"token":        tok,
		"user_id":      user.ID,
		"display_name": user.DisplayName,
	})
}

// --- Teams ---

type teamRow struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Role              string     `json:"role"`
	SubscriptionPlan  string     `json:"subscription_plan"`
	SubscriptionUntil *time.Time `json:"subscription_until"`
	SubscriptionNote  *string    `json:"subscription_note,omitempty"`
}

func (s Service) handleListMyTeams(c *fiber.Ctx) error {
	uid := userID(c)
	rows, err := s.Pool.Query(c.Context(), `
		SELECT t.id, t.name, tm.role, t.subscription_plan, t.subscription_until, t.subscription_note
		FROM team_members tm JOIN teams t ON t.id = tm.team_id
		WHERE tm.user_id = $1 ORDER BY t.created_at`, uid)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	out := []teamRow{}
	for rows.Next() {
		var t teamRow
		if err := rows.Scan(&t.ID, &t.Name, &t.Role, &t.SubscriptionPlan, &t.SubscriptionUntil, &t.SubscriptionNote); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, t)
	}
	return c.JSON(fiber.Map{"teams": out})
}

type createTeamReq struct {
	Name string `json:"name"`
}

func (s Service) handleCreateTeam(c *fiber.Ctx) error {
	uid := userID(c)
	var req createTeamReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > maxTeamNameLen {
		return c.Status(400).JSON(fiber.Map{"error": "name 1..100 chars"})
	}
	isSuper := s.isSuperAdmin(c)

	// Все проверки и INSERT'ы — внутри одной tx с pg_advisory_xact_lock,
	// чтобы устранить TOCTOU между beta-counter и INSERT.
	// Любой concurrent caller встанет в очередь на этот лок до конца нашей tx.
	tx, err := s.Pool.Begin(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	defer tx.Rollback(c.Context())

	if _, err := tx.Exec(c.Context(), "SELECT pg_advisory_xact_lock($1)", teamCreationLockID); err != nil {
		return s.internalErr(c, err)
	}

	// Constraint: 1 owner = 1 company. Super_admin исключён.
	if !isSuper {
		var ownedCount int
		if err := tx.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner'", uid).Scan(&ownedCount); err != nil {
			return s.internalErr(c, err)
		}
		if ownedCount > 0 {
			return c.Status(403).JSON(fiber.Map{
				"error": "you already own a company — one owner = one company in beta",
				"code":  "owner_limit",
			})
		}
	}
	// Beta-лимит: всего N команд в системе. Super_admin обходит.
	if s.BetaTeamLimit > 0 && !isSuper {
		var teamCount int
		if err := tx.QueryRow(c.Context(), "SELECT count(*) FROM teams").Scan(&teamCount); err != nil {
			return s.internalErr(c, err)
		}
		if teamCount >= s.BetaTeamLimit {
			return c.Status(403).JSON(fiber.Map{
				"error": "beta full: free tier limited to " + strconv.Itoa(s.BetaTeamLimit) + " companies",
				"code":  "beta_full",
			})
		}
	}

	teamID := uuid.New()
	if _, err := tx.Exec(c.Context(),
		"INSERT INTO teams (id, name, plan, created_by) VALUES ($1, $2, 'free', $3)",
		teamID, req.Name, uid); err != nil {
		return s.internalErr(c, err)
	}
	if _, err := tx.Exec(c.Context(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'owner')",
		teamID, uid); err != nil {
		return s.internalErr(c, err)
	}
	if err := tx.Commit(c.Context()); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"id": teamID, "name": req.Name, "role": "owner"})
}

func (s Service) handleTeamDetail(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok {
		return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
	}
	var name string
	if err := s.Pool.QueryRow(c.Context(),
		"SELECT name FROM teams WHERE id=$1", teamID).Scan(&name); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "team not found"})
	}
	return c.JSON(fiber.Map{"id": teamID, "name": name, "role": role})
}

func (s Service) handleListMembers(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT u.id, u.email, COALESCE(u.display_name, u.email), tm.role, tm.joined_at
		FROM team_members tm JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1 ORDER BY tm.joined_at`, teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type member struct {
		ID          uuid.UUID `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		Role        string    `json:"role"`
		JoinedAt    time.Time `json:"joined_at"`
	}
	out := []member{}
	for rows.Next() {
		var m member
		if err := rows.Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.JoinedAt); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"members": out})
}

// --- Team summary (per-member ratio за 7 дней) ---

type EventStoreLike interface {
	AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error)
}

var EventStore EventStoreLike

func (s Service) handleTeamSummary(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT u.id, COALESCE(u.display_name, u.email)
		FROM team_members tm JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1 ORDER BY tm.joined_at`, teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type memberStat struct {
		ID          uuid.UUID `json:"id"`
		DisplayName string    `json:"display_name"`
		AIMS        uint64    `json:"ai_ms"`
		ManualMS    uint64    `json:"manual_ms"`
		TotalMS     uint64    `json:"total_ms"`
		AIRatio     int       `json:"ai_ratio"`
	}
	out := []memberStat{}
	memberIDs := []string{}
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	for rows.Next() {
		var m memberStat
		if err := rows.Scan(&m.ID, &m.DisplayName); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, m)
		memberIDs = append(memberIDs, m.ID.String())
	}
	if rows.Err() != nil {
		return s.internalErr(c, rows.Err())
	}
	// Один запрос в EventStore вместо N отдельных по каждому участнику.
	if EventStore != nil && len(memberIDs) > 0 {
		bulk, err := EventStore.AggregateByCategoryBulk(c.Context(), memberIDs, since)
		if err != nil {
			s.Logger.Warn("team summary aggregate failed", zap.Error(err))
		} else {
			for i := range out {
				agg := bulk[out[i].ID.String()]
				out[i].AIMS = agg["ai"]
				out[i].ManualMS = agg["manual"] + agg["refactor"]
				out[i].TotalMS = out[i].AIMS + out[i].ManualMS + agg["other"] + agg["reading"]
				if out[i].TotalMS > 0 {
					out[i].AIRatio = int(float64(out[i].AIMS) * 100.0 / float64(out[i].TotalMS))
				}
			}
		}
	}
	return c.JSON(fiber.Map{"members": out, "since": since})
}

// --- Projects ---

type projectReq struct {
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
}

func (s Service) handleListProjects(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT id, COALESCE(name, repo_url, ''), repo_url, lang_primary, created_at
		FROM projects WHERE team_id = $1 ORDER BY created_at DESC`, teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type project struct {
		ID         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		RepoURL    *string   `json:"repo_url"`
		LangPri    *string   `json:"lang_primary"`
		CreatedAt  time.Time `json:"created_at"`
	}
	out := []project{}
	for rows.Next() {
		var p project
		if err := rows.Scan(&p.ID, &p.Name, &p.RepoURL, &p.LangPri, &p.CreatedAt); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, p)
	}
	return c.JSON(fiber.Map{"projects": out})
}

func (s Service) handleCreateProject(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || (role != "owner" && role != "admin") {
		return c.Status(403).JSON(fiber.Map{"error": "только owner/admin могут создавать проекты"})
	}
	var req projectReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > maxProjectNameLen {
		return c.Status(400).JSON(fiber.Map{"error": "name 1..200 chars"})
	}
	id := uuid.New()
	rootHash := req.RepoURL
	if rootHash == "" {
		rootHash = id.String()
	}
	_, err = s.Pool.Exec(c.Context(), `
		INSERT INTO projects (id, user_id, team_id, name, repo_url, root_path_hash)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)`,
		id, uid, teamID, req.Name, req.RepoURL, rootHash)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"id": id, "name": req.Name})
}

// --- Commits ---

type commitReq struct {
	ProjectID    string `json:"project_id"`
	SHA          string `json:"sha"`
	Message      string `json:"message"`
	Branch       string `json:"branch"`
	FilesChanged int    `json:"files_changed"`
	LinesAdded   int    `json:"lines_added"`
	LinesRemoved int    `json:"lines_removed"`
	AILinesPct   *int   `json:"ai_lines_pct,omitempty"`
	AuthoredAt   string `json:"authored_at"`
}

func (s Service) handleIngestCommit(c *fiber.Ctx) error {
	uid := userID(c)
	var req commitReq
	if err := c.BodyParser(&req); err != nil || req.SHA == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sha required"})
	}
	projID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad project_id"})
	}
	// Достаём team_id проекта и проверяем что юзер в этой команде.
	// Если team_id IS NULL (проект осиротел после удаления команды) — отказываем,
	// иначе любой авторизованный юзер мог бы писать commit'ы в orphan-проект.
	var teamID *uuid.UUID
	if err := s.Pool.QueryRow(c.Context(),
		"SELECT team_id FROM projects WHERE id=$1", projID).Scan(&teamID); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}
	if teamID == nil {
		return c.Status(410).JSON(fiber.Map{"error": "project orphaned (team deleted)"})
	}
	if _, ok := s.teamRole(c.Context(), uid, *teamID); !ok {
		return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
	}
	authoredAt, err := time.Parse(time.RFC3339, req.AuthoredAt)
	if err != nil {
		authoredAt = time.Now().UTC()
	}
	_, err = s.Pool.Exec(c.Context(), `
		INSERT INTO commits (project_id, team_id, user_id, sha, message, branch,
		                     files_changed, lines_added, lines_removed, ai_lines_pct, authored_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (project_id, sha) DO NOTHING`,
		projID, teamID, uid, req.SHA, req.Message, req.Branch,
		req.FilesChanged, req.LinesAdded, req.LinesRemoved, req.AILinesPct, authoredAt)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleProjectCommits(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
	}
	projID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad project id"})
	}
	return s.queryCommits(c, `WHERE project_id = $1 AND team_id = $2 ORDER BY authored_at DESC LIMIT 100`, projID, teamID)
}

func (s Service) handleTeamCommits(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
	}
	return s.queryCommits(c, `WHERE team_id = $1 ORDER BY authored_at DESC LIMIT 100`, teamID)
}

func (s Service) queryCommits(c *fiber.Ctx, where string, args ...any) error {
	rows, err := s.Pool.Query(c.Context(),
		`SELECT c.id, c.project_id, c.user_id, COALESCE(u.display_name, u.email),
		        c.sha, COALESCE(c.message, ''), COALESCE(c.branch, ''),
		        c.files_changed, c.lines_added, c.lines_removed, c.ai_lines_pct, c.authored_at
		 FROM commits c JOIN users u ON u.id = c.user_id `+where, args...)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type commit struct {
		ID            uuid.UUID  `json:"id"`
		ProjectID     *uuid.UUID `json:"project_id"`
		UserID        uuid.UUID  `json:"user_id"`
		Author        string     `json:"author"`
		SHA           string     `json:"sha"`
		Message       string     `json:"message"`
		Branch        string     `json:"branch"`
		FilesChanged  int        `json:"files_changed"`
		LinesAdded    int        `json:"lines_added"`
		LinesRemoved  int        `json:"lines_removed"`
		AILinesPct    *int       `json:"ai_lines_pct"`
		AuthoredAt    time.Time  `json:"authored_at"`
	}
	out := []commit{}
	for rows.Next() {
		var k commit
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.UserID, &k.Author, &k.SHA, &k.Message, &k.Branch,
			&k.FilesChanged, &k.LinesAdded, &k.LinesRemoved, &k.AILinesPct, &k.AuthoredAt); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, k)
	}
	return c.JSON(fiber.Map{"commits": out})
}

// --- Invites ---

type Invite struct {
	ID       uuid.UUID
	TeamID   uuid.UUID
	Code     string
	MaxUses  int
	UseCount int
	Expires  *time.Time
}

func (s Service) findInvite(ctx context.Context, code string) (*Invite, error) {
	var inv Invite
	err := s.Pool.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at
		FROM team_invites
		WHERE code = $1
		  AND (expires_at IS NULL OR expires_at > now())
		  AND use_count < max_uses
		LIMIT 1`, code).Scan(&inv.ID, &inv.TeamID, &inv.Code, &inv.MaxUses, &inv.UseCount, &inv.Expires)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// consumeInvite — атомарно увеличивает use_count если invite ещё валиден.
// Возвращает team_id если получилось (invite ещё не исчерпан и не expired),
// иначе ошибку. Защищает от concurrent accept'ов с одной же ссылки.
func (s Service) consumeInvite(ctx context.Context, code string, userID uuid.UUID) (uuid.UUID, error) {
	var teamID uuid.UUID
	err := s.Pool.QueryRow(ctx, `
		UPDATE team_invites
		SET use_count = use_count + 1, used_by = $2, used_at = now()
		WHERE code = $1
		  AND use_count < max_uses
		  AND (expires_at IS NULL OR expires_at > now())
		RETURNING team_id`, code, userID).Scan(&teamID)
	return teamID, err
}

func (s Service) handleCreateInvite(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad team id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || (role != "owner" && role != "admin") {
		return c.Status(403).JSON(fiber.Map{"error": "только owner/admin могут создавать invites"})
	}
	code := randomCode(16)
	expires := time.Now().Add(7 * 24 * time.Hour)
	_, err = s.Pool.Exec(c.Context(), `
		INSERT INTO team_invites (team_id, code, created_by, max_uses, expires_at)
		VALUES ($1, $2, $3, $4, $5)`, teamID, code, uid, 10, expires)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"code": code, "expires_at": expires, "max_uses": 10})
}

func (s Service) handleInvitePreview(c *fiber.Ctx) error {
	inv, err := s.findInvite(c.Context(), c.Params("code"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "invalid or expired invite"})
	}
	var teamName string
	_ = s.Pool.QueryRow(c.Context(), "SELECT name FROM teams WHERE id=$1", inv.TeamID).Scan(&teamName)
	return c.JSON(fiber.Map{
		"valid":      true,
		"team_id":    inv.TeamID,
		"team_name":  teamName,
		"uses_left":  inv.MaxUses - inv.UseCount,
		"expires_at": inv.Expires,
	})
}

func (s Service) handleInviteAccept(c *fiber.Ctx) error {
	uid := userID(c)
	// consumeInvite атомарно проверяет лимиты и инкрементирует use_count.
	teamID, err := s.consumeInvite(c.Context(), c.Params("code"), uid)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "invalid or expired invite"})
	}
	if err := s.addMember(c.Context(), teamID, uid, "member"); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"team_id": teamID})
}

// --- Super admin ---

// adminListPagination — общий парсер ?limit=N&offset=N для admin-list эндпоинтов.
// limit max=200, default=100. offset >= 0.
func adminListPagination(c *fiber.Ctx) (limit, offset int) {
	limit = 100
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = min(v, 200)
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v > 0 {
		offset = v
	}
	return limit, offset
}

func (s Service) handleAdminListAllTeams(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "super_admin only"})
	}
	limit, offset := adminListPagination(c)
	rows, err := s.Pool.Query(c.Context(), `
		SELECT t.id, t.name, t.plan, t.created_at,
		       t.subscription_plan, t.subscription_until, t.subscription_note,
		       (SELECT count(*) FROM team_members WHERE team_id = t.id) AS member_count,
		       (SELECT u.email FROM team_members tm JOIN users u ON u.id = tm.user_id
		        WHERE tm.team_id = t.id AND tm.role = 'owner' ORDER BY u.created_at LIMIT 1) AS owner_email
		FROM teams t ORDER BY t.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type teamRow struct {
		ID                uuid.UUID  `json:"id"`
		Name              string     `json:"name"`
		Plan              string     `json:"plan"`
		SubscriptionPlan  string     `json:"subscription_plan"`
		SubscriptionUntil *time.Time `json:"subscription_until"`
		SubscriptionNote  *string    `json:"subscription_note"`
		MemberCount       int        `json:"member_count"`
		OwnerEmail        *string    `json:"owner_email"`
		CreatedAt         time.Time  `json:"created_at"`
	}
	out := []teamRow{}
	for rows.Next() {
		var t teamRow
		if err := rows.Scan(&t.ID, &t.Name, &t.Plan, &t.CreatedAt,
			&t.SubscriptionPlan, &t.SubscriptionUntil, &t.SubscriptionNote,
			&t.MemberCount, &t.OwnerEmail); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, t)
	}
	return c.JSON(fiber.Map{"teams": out, "limit": limit, "offset": offset})
}

func (s Service) handleAdminListAllUsers(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "super_admin only"})
	}
	limit, offset := adminListPagination(c)
	rows, err := s.Pool.Query(c.Context(), `
		SELECT u.id, u.email, COALESCE(u.display_name, u.email), u.global_role, u.created_at,
		       (SELECT count(*) FROM team_members WHERE user_id = u.id) AS teams_count
		FROM users u ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type userRow struct {
		ID          uuid.UUID `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		GlobalRole  string    `json:"global_role"`
		CreatedAt   time.Time `json:"created_at"`
		TeamsCount  int       `json:"teams_count"`
	}
	out := []userRow{}
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.GlobalRole, &u.CreatedAt, &u.TeamsCount); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, u)
	}
	return c.JSON(fiber.Map{"users": out, "limit": limit, "offset": offset})
}

// --- helpers ---

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

// --- Beta info ---

func (s Service) handleBetaInfo(c *fiber.Ctx) error {
	var teamCount int
	if err := s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM teams").Scan(&teamCount); err != nil {
		return s.internalErr(c, err)
	}
	limit := s.BetaTeamLimit
	remaining := -1 // -1 = без лимита
	if limit > 0 {
		remaining = max(limit-teamCount, 0)
	}
	return c.JSON(fiber.Map{
		"teams_count":     teamCount,
		"limit":           limit,
		"slots_remaining": remaining,
	})
}

// --- Team update / delete ---

type updateTeamReq struct {
	Name *string `json:"name"`
}

func (s Service) handleUpdateTeam(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || role != "owner" {
		return c.Status(403).JSON(fiber.Map{"error": "только владелец может изменять команду"})
	}
	var req updateTeamReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > maxTeamNameLen {
			return c.Status(400).JSON(fiber.Map{"error": "name 1..100 chars"})
		}
		if _, err := s.Pool.Exec(c.Context(), "UPDATE teams SET name=$1 WHERE id=$2", name, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleDeleteTeam(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || role != "owner" {
		return c.Status(403).JSON(fiber.Map{"error": "только владелец может удалять команду"})
	}
	// Каскад через FK ON DELETE CASCADE удалит team_members, projects, invites, commits.
	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM teams WHERE id=$1", teamID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

// --- Member role / remove ---

type updateMemberRoleReq struct {
	Role string `json:"role"`
}

func (s Service) handleUpdateMemberRole(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	targetUID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || role != "owner" {
		return c.Status(403).JSON(fiber.Map{"error": "только владелец может менять роли"})
	}
	var req updateMemberRoleReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	newRole := strings.ToLower(strings.TrimSpace(req.Role))
	if newRole != "owner" && newRole != "admin" && newRole != "member" {
		return c.Status(400).JSON(fiber.Map{"error": "role must be owner | admin | member"})
	}
	// Если назначаем нового owner — проверка что target не owner'ит другую команду
	// (1-owner-per-user invariant). Super_admin обходит.
	if newRole == "owner" && !s.isSuperAdmin(c) {
		var existingOwned int
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2",
			targetUID, teamID).Scan(&existingOwned)
		if existingOwned > 0 {
			return c.Status(409).JSON(fiber.Map{
				"error": "user already owns another company — one owner = one company in beta",
				"code":  "owner_limit",
			})
		}
	}
	// Если owner понижает себя — не должен быть последним.
	if uid == targetUID && newRole != "owner" {
		var ownerCount int
		_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM team_members WHERE team_id=$1 AND role='owner'", teamID).Scan(&ownerCount)
		if ownerCount <= 1 {
			return c.Status(409).JSON(fiber.Map{"error": "нельзя понизить последнего владельца — назначь другого сначала"})
		}
	}
	if _, err := s.Pool.Exec(c.Context(),
		"UPDATE team_members SET role=$1 WHERE team_id=$2 AND user_id=$3",
		newRole, teamID, targetUID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleRemoveMember(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	targetUID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || (role != "owner" && role != "admin") {
		return c.Status(403).JSON(fiber.Map{"error": "только владелец или админ могут удалять участников"})
	}
	// Нельзя удалить последнего владельца (включая случай "владелец сам себя").
	var targetRole string
	_ = s.Pool.QueryRow(c.Context(),
		"SELECT role FROM team_members WHERE team_id=$1 AND user_id=$2",
		teamID, targetUID).Scan(&targetRole)
	if targetRole == "owner" {
		var ownerCount int
		_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM team_members WHERE team_id=$1 AND role='owner'", teamID).Scan(&ownerCount)
		if ownerCount <= 1 {
			return c.Status(409).JSON(fiber.Map{"error": "нельзя удалить последнего владельца"})
		}
	}
	// Admin не может удалять owner'а.
	if role == "admin" && targetRole == "owner" {
		return c.Status(403).JSON(fiber.Map{"error": "админ не может удалить владельца"})
	}
	if _, err := s.Pool.Exec(c.Context(),
		"DELETE FROM team_members WHERE team_id=$1 AND user_id=$2",
		teamID, targetUID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

// --- Super admin endpoints ---

// requireSuperAdmin — guard для super_admin handler'ов. Возвращает true если пускаем.
func (s Service) requireSuperAdmin(c *fiber.Ctx) bool {
	if !s.isSuperAdmin(c) {
		_ = c.Status(403).JSON(fiber.Map{"error": "super_admin only"})
		return false
	}
	return true
}

func (s Service) handleAdminStats(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	var users, teams, members int
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM users").Scan(&users)
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM teams").Scan(&teams)
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM team_members").Scan(&members)
	return c.JSON(fiber.Map{
		"users_total":   users,
		"teams_total":   teams,
		"members_total": members,
		"beta_limit":    s.BetaTeamLimit,
	})
}

func (s Service) handleAdminDeleteTeam(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM teams WHERE id=$1", teamID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleAdminDeleteUser(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}
	// Запрет на самоудаление, чтобы super_admin не остался без аккаунта.
	if uid == userID(c) {
		return c.Status(409).JSON(fiber.Map{"error": "cannot delete yourself"})
	}
	// Защита: не оставить систему без хотя бы одного super_admin'а.
	var role string
	_ = s.Pool.QueryRow(c.Context(), "SELECT global_role FROM users WHERE id=$1", uid).Scan(&role)
	if role == "super_admin" {
		var count int
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM users WHERE global_role='super_admin'").Scan(&count)
		if count <= 1 {
			return c.Status(409).JSON(fiber.Map{"error": "cannot delete last super_admin"})
		}
	}
	// ClickHouse cleanup ДО Postgres-удаления — после удаления users-row мы потеряем
	// uid, и события юзера останутся orphan'ами (не привязаны к существующему юзеру).
	if d, ok := EventStore.(store.UserDeleter); ok {
		if err := d.DeleteUserData(c.Context(), uid.String()); err != nil {
			s.Logger.Warn("clickhouse delete failed (continuing)",
				zap.String("user_id", uid.String()), zap.Error(err))
		}
	}
	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM users WHERE id=$1", uid); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

type adminUpdateUserReq struct {
	GlobalRole  *string `json:"global_role"`
	DisplayName *string `json:"display_name"`
}

func (s Service) handleAdminUpdateUser(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}
	var req adminUpdateUserReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.GlobalRole != nil {
		role := strings.ToLower(strings.TrimSpace(*req.GlobalRole))
		if role != "user" && role != "super_admin" {
			return c.Status(400).JSON(fiber.Map{"error": "global_role must be user | super_admin"})
		}
		// Запрет понизить себя — иначе можно ребутнуть систему без super_admin'а.
		if uid == userID(c) && role != "super_admin" {
			return c.Status(409).JSON(fiber.Map{"error": "cannot demote yourself"})
		}
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE users SET global_role=$1 WHERE id=$2", role, uid); err != nil {
			return s.internalErr(c, err)
		}
		// Инвалидируем существующие JWT этого юзера — иначе при демоуте super_admin
		// сохранил бы доступ к /v1/admin/* до истечения 14d токена.
		_ = auth.BumpTokenVersion(c.Context(), s.Pool, uid)
	}
	if req.DisplayName != nil {
		dn, ok := validateDisplayName(*req.DisplayName)
		if !ok {
			return c.Status(400).JSON(fiber.Map{"error": "display_name 1..64 chars, no newlines"})
		}
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE users SET display_name=$1 WHERE id=$2", dn, uid); err != nil {
			return s.internalErr(c, err)
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}

type adminAddMemberReq struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (s Service) handleAdminAddMember(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	var req adminAddMemberReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	email, ok := validateEmail(req.Email)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "valid email required"})
	}
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "" {
		role = "member"
	}
	if role != "owner" && role != "admin" && role != "member" {
		return c.Status(400).JSON(fiber.Map{"error": "role must be owner | admin | member"})
	}
	user, err := auth.FindUserByEmail(c.Context(), s.Pool, email)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user with this email not found"})
	}
	// Защита 1-owner invariant: super_admin может назначить owner'а на любую команду,
	// но не на ту, где этот юзер уже owner какой-то другой команды.
	if role == "owner" {
		var existingOwned int
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2",
			user.ID, teamID).Scan(&existingOwned)
		if existingOwned > 0 {
			return c.Status(409).JSON(fiber.Map{
				"error": "user already owns another company",
				"code":  "owner_limit",
			})
		}
	}
	// UPSERT — если юзер уже member, обновим роль (вместо silent no-op).
	if _, err := s.Pool.Exec(c.Context(), `
		INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		teamID, user.ID, role); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true, "user_id": user.ID})
}

// --- Subscriptions (manual billing) ---

type setSubscriptionReq struct {
	Plan  *string `json:"plan"`  // "free" | "pro" | "team" | "enterprise"
	Until *string `json:"until"` // ISO8601 timestamptz; nil = не менять; "" = очистить (revoke)
	Note  *string `json:"note"`  // публичная заметка (видна owner'у)

	// Опциональный payment record. Если есть amount_cents > 0 и covers_until — пишем в team_payments.
	Payment *struct {
		AmountCents int    `json:"amount_cents"`
		Currency    string `json:"currency"`
		Method      string `json:"method"`
		Note        string `json:"note"`
		CoversUntil string `json:"covers_until"` // ISO8601
	} `json:"payment"`
}

func (s Service) handleSetSubscription(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	var req setSubscriptionReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	// Вся валидация — ДО tx.Begin, чтобы плохой запрос не открывал idle txn slot
	// (DoS amplification).
	var planNorm string
	if req.Plan != nil {
		planNorm = strings.ToLower(strings.TrimSpace(*req.Plan))
		if planNorm != "free" && planNorm != "pro" && planNorm != "team" && planNorm != "enterprise" {
			return c.Status(400).JSON(fiber.Map{"error": "plan must be free | pro | team | enterprise"})
		}
	}
	var untilTS *time.Time
	clearUntil := false
	if req.Until != nil {
		if *req.Until == "" {
			clearUntil = true
		} else {
			ts, err := time.Parse(time.RFC3339, *req.Until)
			if err != nil {
				return c.Status(400).JSON(fiber.Map{"error": "until must be ISO8601 (RFC3339)"})
			}
			untilTS = &ts
		}
	}
	var coversUntil *time.Time
	if req.Payment != nil {
		if req.Payment.AmountCents <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "payment.amount_cents must be > 0 if payment is set"})
		}
		ts, err := time.Parse(time.RFC3339, req.Payment.CoversUntil)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "payment.covers_until must be ISO8601"})
		}
		coversUntil = &ts
	}

	tx, err := s.Pool.Begin(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	defer tx.Rollback(c.Context())

	if req.Plan != nil {
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_plan=$1 WHERE id=$2", planNorm, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}
	if clearUntil {
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_until=NULL WHERE id=$1", teamID); err != nil {
			return s.internalErr(c, err)
		}
	} else if untilTS != nil {
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_until=$1 WHERE id=$2", *untilTS, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}
	if req.Note != nil {
		note := strings.TrimSpace(*req.Note)
		var noteVal any
		if note == "" {
			noteVal = nil
		} else {
			noteVal = note
		}
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_note=$1 WHERE id=$2", noteVal, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}

	var paymentID *uuid.UUID
	if coversUntil != nil {
		method := strings.TrimSpace(req.Payment.Method)
		if method == "" {
			method = "manual_transfer"
		}
		currency := strings.ToUpper(strings.TrimSpace(req.Payment.Currency))
		if currency == "" {
			currency = "USD"
		}
		pid := uuid.New()
		if _, err := tx.Exec(c.Context(), `
			INSERT INTO team_payments
			  (id, team_id, amount_cents, currency, method, note, covers_until, recorded_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			pid, teamID, req.Payment.AmountCents, currency, method,
			strings.TrimSpace(req.Payment.Note), *coversUntil, userID(c)); err != nil {
			return s.internalErr(c, err)
		}
		paymentID = &pid
	}

	if err := tx.Commit(c.Context()); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true, "payment_id": paymentID})
}

func (s Service) handleListPayments(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT id, amount_cents, currency, method, note, covers_until, paid_at, recorded_by
		FROM team_payments WHERE team_id=$1 ORDER BY paid_at DESC LIMIT 200`, teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type payment struct {
		ID          uuid.UUID `json:"id"`
		AmountCents int       `json:"amount_cents"`
		Currency    string    `json:"currency"`
		Method      string    `json:"method"`
		Note        string    `json:"note"`
		CoversUntil time.Time `json:"covers_until"`
		PaidAt      time.Time `json:"paid_at"`
		RecordedBy  uuid.UUID `json:"recorded_by"`
	}
	out := []payment{}
	for rows.Next() {
		var p payment
		var note *string
		if err := rows.Scan(&p.ID, &p.AmountCents, &p.Currency, &p.Method, &note,
			&p.CoversUntil, &p.PaidAt, &p.RecordedBy); err != nil {
			return s.internalErr(c, err)
		}
		if note != nil {
			p.Note = *note
		}
		out = append(out, p)
	}
	return c.JSON(fiber.Map{"payments": out})
}
