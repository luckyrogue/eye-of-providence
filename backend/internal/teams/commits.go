// commits.go — git commit ingestion + queries (project-scoped и team-scoped).
package teams

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
