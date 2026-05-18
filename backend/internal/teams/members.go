package teams

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/teams/membersapp"
)

type EventStoreLike interface {
	AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error)
}

func (s Service) handleListMembers(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	if _, ok := s.teamRole(c.Context(), userID(c), teamID); !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	out, err := s.membersApp().List(c.Context(), teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"members": out})
}

func (s Service) handleTeamSummary(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	if _, ok := s.teamRole(c.Context(), userID(c), teamID); !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	sum, err := s.membersApp().TeamSummary(c.Context(), teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	if s.Logger != nil && s.EventStore == nil {
		s.Logger.Warn("team summary: no event store configured")
	}
	type memberStat struct {
		ID          uuid.UUID `json:"id"`
		DisplayName string    `json:"display_name"`
		AIMS        uint64    `json:"ai_ms"`
		ManualMS    uint64    `json:"manual_ms"`
		TotalMS     uint64    `json:"total_ms"`
		AIRatio     int       `json:"ai_ratio"`
	}
	out := make([]memberStat, len(sum.Members))
	for i, m := range sum.Members {
		out[i] = memberStat{
			ID: m.ID, DisplayName: m.DisplayName,
			AIMS: m.AIMS, ManualMS: m.ManualMS, TotalMS: m.TotalMS, AIRatio: m.AIRatio,
		}
	}
	return c.JSON(fiber.Map{"members": out, "since": sum.Since})
}

type updateMemberRoleReq struct {
	Role string `json:"role"`
}

func (s Service) handleUpdateMemberRole(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	targetUID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_user_id", "invalid user id")
	}
	var req updateMemberRoleReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	err = s.membersApp().UpdateRole(c.Context(), membersapp.UpdateRoleInput{
		ActorID: userID(c), TeamID: teamID, TargetUserID: targetUID,
		NewRole: req.Role, AllowOwner: s.isSuperAdmin(c),
	})
	if err != nil {
		switch {
		case errors.Is(err, membersapp.ErrOwnerRequired):
			return httperr.Forbidden(c, "owner_required", "only owner can change roles")
		case errors.Is(err, membersapp.ErrInvalidRole):
			return httperr.BadRequest(c, "invalid_role", "role must be owner | admin | member")
		case errors.Is(err, membersapp.ErrOwnerLimit):
			return httperr.Conflict(c, "owner_limit", "user already owns another company — beta limits to 1 owner = 1 company")
		case errors.Is(err, membersapp.ErrLastOwner):
			return httperr.Conflict(c, "last_owner", "cannot demote last owner — assign another first")
		default:
			return s.internalErr(c, err)
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleRemoveMember(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	targetUID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_user_id", "invalid user id")
	}
	err = s.membersApp().Remove(c.Context(), membersapp.RemoveMemberInput{
		ActorID: userID(c), TeamID: teamID, TargetUserID: targetUID,
	})
	if err != nil {
		switch {
		case errors.Is(err, membersapp.ErrRoleInsufficient):
			return httperr.Forbidden(c, "role_insufficient", "only owner/admin can remove members")
		case errors.Is(err, membersapp.ErrLastOwner):
			return httperr.Conflict(c, "last_owner", "cannot remove last owner")
		case errors.Is(err, membersapp.ErrAdminCantRemove):
			return httperr.Forbidden(c, "admin_cant_remove_owner", "admin cannot remove owner")
		default:
			return s.internalErr(c, err)
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}
