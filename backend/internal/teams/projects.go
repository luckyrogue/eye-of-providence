// projects.go — список + создание project'ов команды.
package teams

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		RepoURL   *string   `json:"repo_url"`
		LangPri   *string   `json:"lang_primary"`
		CreatedAt time.Time `json:"created_at"`
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
