package teams

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
)

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
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > maxTeamNameLen {
		return httperr.BadRequest(c, "invalid_team_name", "name must be 1..100 chars")
	}
	isSuper := s.isSuperAdmin(c)

	tx, err := s.Pool.Begin(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	defer tx.Rollback(c.Context())

	if _, err := tx.Exec(c.Context(), "SELECT pg_advisory_xact_lock($1)", teamCreationLockID); err != nil {
		return s.internalErr(c, err)
	}

	if !isSuper {
		var ownedCount int
		if err := tx.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner'", uid).Scan(&ownedCount); err != nil {
			return s.internalErr(c, err)
		}
		if ownedCount > 0 {
			return httperr.Forbidden(c, "owner_limit", "already an owner — beta limits to 1 owner = 1 company")
		}
	}

	if s.BetaTeamLimit > 0 && !isSuper {
		var teamCount int
		if err := tx.QueryRow(c.Context(), "SELECT count(*) FROM teams").Scan(&teamCount); err != nil {
			return s.internalErr(c, err)
		}
		if teamCount >= s.BetaTeamLimit {
			return httperr.Forbidden(c, "beta_full", "beta full: free tier limited to "+strconv.Itoa(s.BetaTeamLimit)+" companies")
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
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok {
		return httperr.Forbidden(c, "not_member", "not a team member")
	}
	var name string
	if err := s.Pool.QueryRow(c.Context(),
		"SELECT name FROM teams WHERE id=$1", teamID).Scan(&name); err != nil {
		return httperr.NotFound(c, "team_not_found", "team not found")
	}
	return c.JSON(fiber.Map{"id": teamID, "name": name, "role": role})
}

type updateTeamReq struct {
	Name *string `json:"name"`
}

func (s Service) handleUpdateTeam(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || role != "owner" {
		return httperr.Forbidden(c, "owner_required", "only owner can modify team")
	}
	var req updateTeamReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > maxTeamNameLen {
			return httperr.BadRequest(c, "invalid_team_name", "name must be 1..100 chars")
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
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || role != "owner" {
		return httperr.Forbidden(c, "owner_required", "only owner can delete team")
	}

	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM teams WHERE id=$1", teamID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleBetaInfo(c *fiber.Ctx) error {
	var teamCount int
	if err := s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM teams").Scan(&teamCount); err != nil {
		return s.internalErr(c, err)
	}
	limit := s.BetaTeamLimit
	remaining := -1
	if limit > 0 {
		remaining = max(limit-teamCount, 0)
	}
	return c.JSON(fiber.Map{
		"teams_count":     teamCount,
		"limit":           limit,
		"slots_remaining": remaining,
	})
}
