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
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
)

const (
	minPasswordLen    = 8
	maxPasswordLen    = 256
	maxDisplayNameLen = 64
	maxTeamNameLen    = 100
	maxProjectNameLen = 200
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

const tokenTTL = 90 * 24 * time.Hour

type Service struct {
	Pool       *pgxpool.Pool
	JWTSecret  string
	Logger     *zap.Logger
	InviteOnly bool // регистрация только по invite (первый user всегда может — bootstrap)
}

func RegisterRoutes(app *fiber.App, s Service) {
	// Public
	a := app.Group("/v1/auth")
	a.Post("/register", s.handleRegister)
	a.Post("/login", s.handleLogin)
	a.Get("/config", s.handleAuthConfig)
	app.Get("/v1/invites/:code", s.handleInvitePreview)

	// Authed
	g := app.Group("/v1", auth.Middleware(s.JWTSecret))

	g.Get("/teams", s.handleListMyTeams)
	g.Post("/teams", s.handleCreateTeam)

	t := g.Group("/teams/:id")
	t.Get("/", s.handleTeamDetail)
	t.Get("/members", s.handleListMembers)
	t.Post("/invites", s.handleCreateInvite)
	t.Get("/summary", s.handleTeamSummary)

	t.Get("/projects", s.handleListProjects)
	t.Post("/projects", s.handleCreateProject)
	t.Get("/projects/:project_id/commits", s.handleProjectCommits)

	t.Get("/commits", s.handleTeamCommits)
	g.Post("/commits", s.handleIngestCommit) // user pushes commit info через CLI/git hook

	g.Post("/invites/:code/accept", s.handleInviteAccept)

	// Super admin
	g.Get("/admin/teams", s.handleAdminListAllTeams)
	g.Get("/admin/users", s.handleAdminListAllUsers)
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

	var joinedTeam *uuid.UUID
	if req.InviteCode != nil && *req.InviteCode != "" {
		inv, err := s.findInvite(c.Context(), *req.InviteCode)
		if err == nil {
			if err := s.addMember(c.Context(), inv.TeamID, user.ID, "member"); err == nil {
				_ = s.consumeInvite(c.Context(), inv.Code, user.ID)
				joinedTeam = &inv.TeamID
			}
		}
	}

	tok, err := auth.IssueJWT(s.JWTSecret, user.ID.String(), user.Email, "password", tokenTTL)
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
	tok, err := auth.IssueJWT(s.JWTSecret, user.ID.String(), user.Email, "password", tokenTTL)
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
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role string    `json:"role"`
}

func (s Service) handleListMyTeams(c *fiber.Ctx) error {
	uid := userID(c)
	rows, err := s.Pool.Query(c.Context(), `
		SELECT t.id, t.name, tm.role
		FROM team_members tm JOIN teams t ON t.id = tm.team_id
		WHERE tm.user_id = $1 ORDER BY t.created_at`, uid)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	out := []teamRow{}
	for rows.Next() {
		var t teamRow
		if err := rows.Scan(&t.ID, &t.Name, &t.Role); err != nil {
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
	tx, err := s.Pool.Begin(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	defer tx.Rollback(c.Context())
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
	AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)
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
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	for rows.Next() {
		var m memberStat
		if err := rows.Scan(&m.ID, &m.DisplayName); err != nil {
			return s.internalErr(c, err)
		}
		if EventStore != nil {
			agg, _ := EventStore.AggregateByCategory(c.Context(), m.ID.String(), since)
			m.AIMS = agg["ai"]
			m.ManualMS = agg["manual"] + agg["refactor"]
			m.TotalMS = m.AIMS + m.ManualMS + agg["other"] + agg["reading"]
			if m.TotalMS > 0 {
				m.AIRatio = int(float64(m.AIMS) * 100.0 / float64(m.TotalMS))
			}
		}
		out = append(out, m)
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
	// Достаём team_id проекта и проверяем что юзер в этой команде
	var teamID *uuid.UUID
	if err := s.Pool.QueryRow(c.Context(),
		"SELECT team_id FROM projects WHERE id=$1", projID).Scan(&teamID); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}
	if teamID != nil {
		if _, ok := s.teamRole(c.Context(), uid, *teamID); !ok {
			return c.Status(403).JSON(fiber.Map{"error": "not a team member"})
		}
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

func (s Service) consumeInvite(ctx context.Context, code string, userID uuid.UUID) error {
	_, err := s.Pool.Exec(ctx, `
		UPDATE team_invites
		SET use_count = use_count + 1, used_by = $2, used_at = now()
		WHERE code = $1`, code, userID)
	return err
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
	inv, err := s.findInvite(c.Context(), c.Params("code"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "invalid or expired invite"})
	}
	if err := s.addMember(c.Context(), inv.TeamID, uid, "member"); err != nil {
		return s.internalErr(c, err)
	}
	_ = s.consumeInvite(c.Context(), inv.Code, uid)
	return c.JSON(fiber.Map{"team_id": inv.TeamID})
}

// --- Super admin ---

func (s Service) handleAdminListAllTeams(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "super_admin only"})
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT t.id, t.name, t.plan, t.created_at,
		       (SELECT count(*) FROM team_members WHERE team_id = t.id) AS member_count
		FROM teams t ORDER BY t.created_at DESC`)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type teamRow struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Plan      string    `json:"plan"`
		Members   int       `json:"members"`
		CreatedAt time.Time `json:"created_at"`
	}
	out := []teamRow{}
	for rows.Next() {
		var t teamRow
		if err := rows.Scan(&t.ID, &t.Name, &t.Plan, &t.CreatedAt, &t.Members); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, t)
	}
	return c.JSON(fiber.Map{"teams": out})
}

func (s Service) handleAdminListAllUsers(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "super_admin only"})
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT id, email, COALESCE(display_name, email), global_role, created_at
		FROM users ORDER BY created_at DESC LIMIT 200`)
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
	}
	out := []userRow{}
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.GlobalRole, &u.CreatedAt); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, u)
	}
	return c.JSON(fiber.Map{"users": out})
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
