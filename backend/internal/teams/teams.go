package teams

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/teams/teamsapp"
)

func (s Service) handleListMyTeams(c *fiber.Ctx) error {
	out, err := s.teamsApp().ListMine(c.Context(), userID(c))
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"teams": out})
}

type createTeamReq struct {
	Name string `json:"name"`
}

func (s Service) handleCreateTeam(c *fiber.Ctx) error {
	var req createTeamReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	out, err := s.teamsApp().Create(c.Context(), teamsapp.CreateInput{
		UserID: userID(c), Name: req.Name, IsSuper: s.isSuperAdmin(c),
		BetaLimit: s.BetaTeamLimit, LockID: teamCreationLockID,
	})
	if err != nil {
		switch {
		case errors.Is(err, teamsapp.ErrInvalidName):
			return httperr.BadRequest(c, "invalid_team_name", "name must be 1..100 chars")
		case errors.Is(err, teamsapp.ErrOwnerLimit):
			return httperr.Forbidden(c, "owner_limit", "already an owner — beta limits to 1 owner = 1 company")
		case errors.Is(err, teamsapp.ErrBetaFull):
			return httperr.Forbidden(c, "beta_full", "beta full: free tier limited to "+strconv.Itoa(s.BetaTeamLimit)+" companies")
		default:
			return s.internalErr(c, err)
		}
	}
	return c.JSON(fiber.Map{"id": out.ID, "name": out.Name, "role": out.Role})
}

func (s Service) handleTeamDetail(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	role, ok := s.teamRole(c.Context(), userID(c), teamID)
	if !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	detail, err := s.teamsApp().GetDetail(c.Context(), userID(c), teamID, role)
	if err != nil {
		if errors.Is(err, teamsapp.ErrTeamNotFound) {
			return httperr.NotFound(c, "team_not_found", "team not found")
		}
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"id": detail.ID, "name": detail.Name, "role": detail.Role})
}

type updateTeamReq struct {
	Name *string `json:"name"`
}

func (s Service) handleUpdateTeam(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	role, ok := s.teamRole(c.Context(), userID(c), teamID)
	if !ok || role != "owner" {
		return httperr.Forbidden(c, "owner_required", "only owner can modify team")
	}
	var req updateTeamReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	if req.Name != nil {
		if err := s.teamsApp().Update(c.Context(), teamID, *req.Name); err != nil {
			if errors.Is(err, teamsapp.ErrInvalidName) {
				return httperr.BadRequest(c, "invalid_team_name", "name must be 1..100 chars")
			}
			return s.internalErr(c, err)
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleDeleteTeam(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	role, ok := s.teamRole(c.Context(), userID(c), teamID)
	if !ok || role != "owner" {
		return httperr.Forbidden(c, "owner_required", "only owner can delete team")
	}
	if err := s.teamsApp().Delete(c.Context(), teamID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleBetaInfo(c *fiber.Ctx) error {
	info, err := s.teamsApp().BetaInfo(c.Context(), s.BetaTeamLimit)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"teams_count":     info.TeamsCount,
		"limit":           info.Limit,
		"slots_remaining": info.SlotsRemaining,
	})
}
