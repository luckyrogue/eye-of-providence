package teams

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/commitsapp"
	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type commitIngestAdapter struct{ s *Service }

func (a commitIngestAdapter) Ingest(ctx context.Context, userID uuid.UUID, in domain.CommitInput) (bool, uuid.UUID, uuid.UUID, error) {
	projID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return false, uuid.Nil, uuid.Nil, err
	}
	teamID, err := commitsapp.NewPGProjectTeams(a.s.Pool).TeamIDForProject(ctx, projID)
	if err != nil {
		return false, uuid.Nil, uuid.Nil, err
	}
	tag, err := a.s.Pool.Exec(ctx, `
		INSERT INTO commits (project_id, team_id, user_id, sha, message, branch,
		                     files_changed, lines_added, lines_removed, ai_lines_pct, authored_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (project_id, sha) DO NOTHING`,
		projID, teamID, userID, in.SHA, in.Message, in.Branch,
		in.FilesChanged, in.LinesAdded, in.LinesRemoved, in.AILinesPct, in.AuthoredAt)
	if err != nil {
		return false, uuid.Nil, uuid.Nil, err
	}
	inserted := tag.RowsAffected() > 0
	if inserted && a.s.Webhooks != nil {
		req := commitReq{
			ProjectID: in.ProjectID, SHA: in.SHA, Message: in.Message, Branch: in.Branch,
			FilesChanged: in.FilesChanged, LinesAdded: in.LinesAdded, LinesRemoved: in.LinesRemoved,
		}
		dispatchCommitWebhook(ctx, *a.s, userID, projID, *teamID, req, in.AuthoredAt)
	}
	return inserted, *teamID, projID, nil
}

func (s *Service) commitsApp() *commitsapp.Service {
	return commitsapp.New(commitsapp.Deps{
		Ingest:   commitIngestAdapter{s: s},
		Commits:  commitsapp.NewPGCommits(s.Pool),
		Projects: commitsapp.NewPGProjectTeams(s.Pool),
		Roles:    memberRoleAdapter{s: s},
	})
}
