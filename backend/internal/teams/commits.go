package teams

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/teams/domain"
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
	var req commitReq
	if err := c.BodyParser(&req); err != nil || req.SHA == "" {
		return httperr.BadRequest(c, "sha_required", "sha required")
	}
	authoredAt, err := time.Parse(time.RFC3339, req.AuthoredAt)
	if err != nil {
		authoredAt = time.Now().UTC()
	}
	_, err = s.commitsApp().IngestForMember(c.Context(), userID(c), domain.CommitInput{
		ProjectID: req.ProjectID, SHA: req.SHA, Message: req.Message, Branch: req.Branch,
		FilesChanged: req.FilesChanged, LinesAdded: req.LinesAdded, LinesRemoved: req.LinesRemoved,
		AILinesPct: req.AILinesPct, AuthoredAt: authoredAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return httperr.Forbidden(c, "not_member", "not a team member")
		}
		if errors.Is(err, domain.ErrProjectNotFound) {
			return httperr.NotFound(c, "project_not_found", "project not found")
		}
		if errors.Is(err, domain.ErrProjectOrphaned) {
			return httperr.Gone(c, "project_orphaned", "project orphaned (team deleted)")
		}
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func dispatchCommitWebhook(ctx context.Context, s Service, userID, projectID, teamID uuid.UUID, req commitReq, authoredAt time.Time) {
	_ = ctx
	s.Webhooks.Dispatch(userID, "commit.ingested", map[string]any{
		"user_id": userID, "project_id": projectID, "team_id": teamID,
		"sha": req.SHA, "message": req.Message, "branch": req.Branch,
		"files_changed": req.FilesChanged, "lines_added": req.LinesAdded,
		"lines_removed": req.LinesRemoved,
		"authored_at":   authoredAt.UTC().Format(time.RFC3339),
	})
}

func (s Service) handleProjectCommits(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	projID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_project_id", "bad project id")
	}
	rows, err := s.commitsApp().ListByProject(c.Context(), userID(c), teamID, projID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return httperr.Forbidden(c, "not_member", "not a team member")
		}
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"commits": rows})
}

func (s Service) handleTeamCommits(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	rows, err := s.commitsApp().ListByTeam(c.Context(), userID(c), teamID)
	if err != nil {
		if errors.Is(err, domain.ErrNotMember) {
			return httperr.Forbidden(c, "not_member", "not a team member")
		}
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"commits": rows})
}
