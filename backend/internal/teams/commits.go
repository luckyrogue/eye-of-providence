package teams

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
)

type WebhookDispatcher interface {
	Dispatch(userID uuid.UUID, event string, payload any)
}

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
		return httperr.BadRequest(c, "sha_required", "sha required")
	}
	projID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return httperr.BadRequest(c, "invalid_project_id", "bad project_id")
	}

	var teamID *uuid.UUID
	if err := s.Pool.QueryRow(c.Context(),
		"SELECT team_id FROM projects WHERE id=$1", projID).Scan(&teamID); err != nil {
		return httperr.NotFound(c, "project_not_found", "project not found")
	}
	if teamID == nil {
		return httperr.Gone(c, "project_orphaned", "project orphaned (team deleted)")
	}
	if _, ok := s.teamRole(c.Context(), uid, *teamID); !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	authoredAt, err := time.Parse(time.RFC3339, req.AuthoredAt)
	if err != nil {
		authoredAt = time.Now().UTC()
	}
	tag, err := s.Pool.Exec(c.Context(), `
		INSERT INTO commits (project_id, team_id, user_id, sha, message, branch,
		                     files_changed, lines_added, lines_removed, ai_lines_pct, authored_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (project_id, sha) DO NOTHING`,
		projID, teamID, uid, req.SHA, req.Message, req.Branch,
		req.FilesChanged, req.LinesAdded, req.LinesRemoved, req.AILinesPct, authoredAt)
	if err != nil {
		return s.internalErr(c, err)
	}

	if tag.RowsAffected() > 0 && s.Webhooks != nil {
		dispatchCommitWebhook(c.Context(), s, uid, projID, *teamID, req, authoredAt)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func dispatchCommitWebhook(ctx context.Context, s Service, userID, projectID, teamID uuid.UUID, req commitReq, authoredAt time.Time) {
	_ = ctx
	s.Webhooks.Dispatch(userID, "commit.ingested", map[string]any{
		"user_id":       userID,
		"project_id":    projectID,
		"team_id":       teamID,
		"sha":           req.SHA,
		"message":       req.Message,
		"branch":        req.Branch,
		"files_changed": req.FilesChanged,
		"lines_added":   req.LinesAdded,
		"lines_removed": req.LinesRemoved,
		"authored_at":   authoredAt.UTC().Format(time.RFC3339),
	})
}

func (s Service) handleProjectCommits(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	projID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_project_id", "bad project id")
	}
	return s.queryCommits(c, `WHERE c.project_id = $1 AND c.team_id = $2 ORDER BY c.authored_at DESC LIMIT 100`, projID, teamID)
}

func (s Service) handleTeamCommits(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	return s.queryCommits(c, `WHERE c.team_id = $1 ORDER BY c.authored_at DESC LIMIT 100`, teamID)
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
		ID           uuid.UUID  `json:"id"`
		ProjectID    *uuid.UUID `json:"project_id"`
		UserID       uuid.UUID  `json:"user_id"`
		Author       string     `json:"author"`
		SHA          string     `json:"sha"`
		Message      string     `json:"message"`
		Branch       string     `json:"branch"`
		FilesChanged int        `json:"files_changed"`
		LinesAdded   int        `json:"lines_added"`
		LinesRemoved int        `json:"lines_removed"`
		AILinesPct   *int       `json:"ai_lines_pct"`
		AuthoredAt   time.Time  `json:"authored_at"`
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
