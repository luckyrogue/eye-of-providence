package teams

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
)

type projectReq struct {
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
}

func (s Service) handleListProjects(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	if _, ok := s.teamRole(c.Context(), uid, teamID); !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
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
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || (role != "owner" && role != "admin") {
		return httperr.Forbidden(c, "role_insufficient", "only owner/admin can create projects")
	}
	var req projectReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > maxProjectNameLen {
		return httperr.BadRequest(c, "invalid_project_name", "name must be 1..200 chars")
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
