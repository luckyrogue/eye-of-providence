package teams

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/teams/domain"
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
	out, err := s.projectsApp().List(c.Context(), teamID)
	if err != nil {
		return s.internalErr(c, err)
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
	id, err := s.projectsApp().Create(c.Context(), teamID, uid, domain.CreateProjectInput{
		Name: req.Name, RepoURL: req.RepoURL,
	})
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"id": id, "name": req.Name})
}
