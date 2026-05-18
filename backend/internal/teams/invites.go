package teams

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/teams/invitesapp"
)

func (s Service) handleCreateInvite(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	var req struct {
		Email string `json:"email"`
	}
	_ = c.BodyParser(&req)
	out, err := s.invitesApp().Create(c.Context(), invitesapp.CreateInput{
		TeamID: teamID, CreatedBy: userID(c), Email: req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, invitesapp.ErrRoleInsufficient):
			return httperr.Forbidden(c, "role_insufficient", "only owner/admin can create invites")
		case errors.Is(err, invitesapp.ErrInvalidEmail):
			return httperr.BadRequest(c, "invalid_email", "invalid email")
		default:
			return s.internalErr(c, err)
		}
	}
	resp := fiber.Map{"code": out.Code, "expires_at": out.ExpiresAt, "max_uses": out.MaxUses}
	if out.Email != "" {
		resp["email"] = out.Email
		resp["sent"] = out.Sent
	}
	return c.JSON(resp)
}

func (s Service) handleInvitePreview(c *fiber.Ctx) error {
	prev, err := s.invitesApp().Preview(c.Context(), c.Params("code"))
	if err != nil {
		return httperr.NotFound(c, "invite_invalid", "invalid or expired invite")
	}
	return c.JSON(fiber.Map{
		"valid": true, "team_id": prev.TeamID, "team_name": prev.TeamName,
		"uses_left": prev.UsesLeft, "expires_at": prev.ExpiresAt,
	})
}

func (s Service) handleInviteAccept(c *fiber.Ctx) error {
	teamID, err := s.invitesApp().Accept(c.Context(), c.Params("code"), userID(c))
	if err != nil {
		if errors.Is(err, invitesapp.ErrPlanLimitExceeded) {
			return httperr.Forbidden(c, "plan_limit_exceeded", err.Error())
		}
		return httperr.NotFound(c, "invite_invalid", "invalid or expired invite")
	}
	return c.JSON(fiber.Map{"team_id": teamID})
}
