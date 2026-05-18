package teams

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
	"github.com/eye-of-providence/backend/internal/teams/projectsapp"
)

type projectRepoAdapter struct{ s *Service }

func (a projectRepoAdapter) List(ctx context.Context, teamID uuid.UUID) ([]domain.Project, error) {
	rows, err := a.s.Pool.Query(ctx, `
		SELECT id, COALESCE(name, repo_url, ''), repo_url, lang_primary, created_at
		FROM projects WHERE team_id = $1 ORDER BY created_at DESC`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Project{}
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.RepoURL, &p.LangPri, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (a projectRepoAdapter) Create(ctx context.Context, teamID, userID uuid.UUID, in domain.CreateProjectInput) (uuid.UUID, error) {
	id := uuid.New()
	rootHash := in.RepoURL
	if rootHash == "" {
		rootHash = id.String()
	}
	_, err := a.s.Pool.Exec(ctx, `
		INSERT INTO projects (id, user_id, team_id, name, repo_url, root_path_hash)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)`,
		id, userID, teamID, strings.TrimSpace(in.Name), in.RepoURL, rootHash)
	return id, err
}

func (s *Service) projectsApp() *projectsapp.Service {
	return projectsapp.New(projectsapp.Deps{Repo: projectRepoAdapter{s: s}})
}
