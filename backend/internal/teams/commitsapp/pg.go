package commitsapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type pgCommits struct {
	pool *pgxpool.Pool
}

func NewPGCommits(pool *pgxpool.Pool) CommitReader {
	return pgCommits{pool: pool}
}

func (p pgCommits) query(ctx context.Context, where string, args ...any) ([]CommitRow, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT c.id, c.project_id, c.user_id, COALESCE(u.display_name, u.email),
		       c.sha, COALESCE(c.message, ''), COALESCE(c.branch, ''),
		       c.files_changed, c.lines_added, c.lines_removed, c.ai_lines_pct, c.authored_at
		 FROM commits c JOIN users u ON u.id = c.user_id `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CommitRow{}
	for rows.Next() {
		var k CommitRow
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.UserID, &k.Author, &k.SHA, &k.Message, &k.Branch,
			&k.FilesChanged, &k.LinesAdded, &k.LinesRemoved, &k.AILinesPct, &k.AuthoredAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (p pgCommits) ListByProject(ctx context.Context, teamID, projectID uuid.UUID) ([]CommitRow, error) {
	return p.query(ctx, `WHERE c.project_id = $1 AND c.team_id = $2 ORDER BY c.authored_at DESC LIMIT 100`, projectID, teamID)
}

func (p pgCommits) ListByTeam(ctx context.Context, teamID uuid.UUID) ([]CommitRow, error) {
	return p.query(ctx, `WHERE c.team_id = $1 ORDER BY c.authored_at DESC LIMIT 100`, teamID)
}

type pgProjectTeams struct {
	pool *pgxpool.Pool
}

func NewPGProjectTeams(pool *pgxpool.Pool) ProjectTeamResolver {
	return pgProjectTeams{pool: pool}
}

func (p pgProjectTeams) TeamIDForProject(ctx context.Context, projectID uuid.UUID) (*uuid.UUID, error) {
	var teamID *uuid.UUID
	err := p.pool.QueryRow(ctx, `SELECT team_id FROM projects WHERE id=$1`, projectID).Scan(&teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}
	if teamID == nil {
		return nil, domain.ErrProjectOrphaned
	}
	return teamID, nil
}
