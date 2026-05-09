// members.go — list / summary / role-update / remove членов команды.
package teams

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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
	AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error)
}

// EventStore — синглтон, в который cmd/api инжектит ClickHouse store.
// Объявлен в pkg-level чтобы handler'ы могли его читать без передачи через Service.
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
	memberIDs := []string{}
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	for rows.Next() {
		var m memberStat
		if err := rows.Scan(&m.ID, &m.DisplayName); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, m)
		memberIDs = append(memberIDs, m.ID.String())
	}
	if rows.Err() != nil {
		return s.internalErr(c, rows.Err())
	}
	// Один запрос в EventStore вместо N отдельных по каждому участнику.
	if EventStore != nil && len(memberIDs) > 0 {
		bulk, err := EventStore.AggregateByCategoryBulk(c.Context(), memberIDs, since)
		if err != nil {
			s.Logger.Warn("team summary aggregate failed", zap.Error(err))
		} else {
			for i := range out {
				agg := bulk[out[i].ID.String()]
				out[i].AIMS = agg["ai"]
				out[i].ManualMS = agg["manual"] + agg["refactor"]
				out[i].TotalMS = out[i].AIMS + out[i].ManualMS + agg["other"] + agg["reading"]
				if out[i].TotalMS > 0 {
					out[i].AIRatio = int(float64(out[i].AIMS) * 100.0 / float64(out[i].TotalMS))
				}
			}
		}
	}
	return c.JSON(fiber.Map{"members": out, "since": since})
}

// --- Role update / remove ---

type updateMemberRoleReq struct {
	Role string `json:"role"`
}

func (s Service) handleUpdateMemberRole(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	targetUID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || role != "owner" {
		return c.Status(403).JSON(fiber.Map{"error": "только владелец может менять роли"})
	}
	var req updateMemberRoleReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	newRole := strings.ToLower(strings.TrimSpace(req.Role))
	if newRole != "owner" && newRole != "admin" && newRole != "member" {
		return c.Status(400).JSON(fiber.Map{"error": "role must be owner | admin | member"})
	}
	// Если назначаем нового owner — проверка что target не owner'ит другую команду
	// (1-owner-per-user invariant). Super_admin обходит.
	if newRole == "owner" && !s.isSuperAdmin(c) {
		var existingOwned int
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2",
			targetUID, teamID).Scan(&existingOwned)
		if existingOwned > 0 {
			return c.Status(409).JSON(fiber.Map{
				"error": "пользователь уже владелец другой компании — в бете 1 owner = 1 company",
				"code":  "owner_limit",
			})
		}
	}
	// Если owner понижает себя — не должен быть последним.
	if uid == targetUID && newRole != "owner" {
		var ownerCount int
		_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM team_members WHERE team_id=$1 AND role='owner'", teamID).Scan(&ownerCount)
		if ownerCount <= 1 {
			return c.Status(409).JSON(fiber.Map{"error": "нельзя понизить последнего владельца — назначь другого сначала"})
		}
	}
	if _, err := s.Pool.Exec(c.Context(),
		"UPDATE team_members SET role=$1 WHERE team_id=$2 AND user_id=$3",
		newRole, teamID, targetUID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleRemoveMember(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid team id"})
	}
	targetUID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || (role != "owner" && role != "admin") {
		return c.Status(403).JSON(fiber.Map{"error": "только владелец или админ могут удалять участников"})
	}
	// Нельзя удалить последнего владельца (включая случай "владелец сам себя").
	var targetRole string
	_ = s.Pool.QueryRow(c.Context(),
		"SELECT role FROM team_members WHERE team_id=$1 AND user_id=$2",
		teamID, targetUID).Scan(&targetRole)
	if targetRole == "owner" {
		var ownerCount int
		_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM team_members WHERE team_id=$1 AND role='owner'", teamID).Scan(&ownerCount)
		if ownerCount <= 1 {
			return c.Status(409).JSON(fiber.Map{"error": "нельзя удалить последнего владельца"})
		}
	}
	// Admin не может удалять owner'а.
	if role == "admin" && targetRole == "owner" {
		return c.Status(403).JSON(fiber.Map{"error": "админ не может удалить владельца"})
	}
	if _, err := s.Pool.Exec(c.Context(),
		"DELETE FROM team_members WHERE team_id=$1 AND user_id=$2",
		teamID, targetUID); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}
