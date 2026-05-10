// Package prcomment — постит комментарии в PR/MR с aggregate AI-attribution
// данными по commits. Pull-based (user owns provider token):
//
//	POST /v1/integrations/pr-comment
//	  Authorization: Bearer <eop_token>
//	  body: {
//	    provider: "github" | "gitlab",
//	    host:     "https://gitlab.example.com" (для self-hosted GL; GH default api.github.com)
//	    repo:     "owner/name" (GitHub) | "namespace/project" (GitLab)
//	    pr_number: int (GitHub PR | GitLab MR IID)
//	    shas:     ["abc1234", ...] commits в PR
//	    provider_token: "ghp_xxx" / "glpat-xxx"
//	  }
package prcomment

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Aggregate — итоговые числа по PR/MR. Frontend / Slack / comment-formatter
// все используют один shape.
type Aggregate struct {
	TotalCommits   int    `json:"total_commits"`
	WithAttribution int   `json:"with_attribution"` // commits где ai_lines_pct != null
	LinesAdded     int    `json:"lines_added"`
	LinesRemoved   int    `json:"lines_removed"`
	// AIPercentWeighted — weighted average ai_lines_pct по lines_added.
	// nil если все commits без attribution (агент не установлен / скан ещё не
	// прошёл).
	AIPercentWeighted *float64 `json:"ai_percent,omitempty"`
}

// AggregateBySHA — query commits по списку SHA для конкретного user'а
// (auth-context). Скоп — last 100 commits через user_id ownership.
//
// Defensive: SHA matching сделан через ANY($1) чтобы избежать SQL-injection
// + дешёвой памяти. Дубликаты в shas игнорируются (DISTINCT не нужен —
// commits уникальны по (project_id, sha)).
func AggregateBySHA(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, shas []string) (Aggregate, error) {
	if len(shas) == 0 {
		return Aggregate{}, nil
	}
	row := pool.QueryRow(ctx, `
		SELECT
		  COUNT(*) AS total_commits,
		  COUNT(*) FILTER (WHERE c.ai_lines_pct IS NOT NULL) AS with_attribution,
		  COALESCE(SUM(c.lines_added), 0) AS lines_added,
		  COALESCE(SUM(c.lines_removed), 0) AS lines_removed,
		  -- weighted avg: sum(pct * lines) / sum(lines), только rows с pct.
		  -- NULLIF защищает от divide-by-zero когда все commits 0-line.
		  CASE
		    WHEN COALESCE(SUM(c.lines_added) FILTER (WHERE c.ai_lines_pct IS NOT NULL), 0) > 0
		    THEN SUM((c.ai_lines_pct * c.lines_added)::numeric) FILTER (WHERE c.ai_lines_pct IS NOT NULL)
		         / NULLIF(SUM(c.lines_added) FILTER (WHERE c.ai_lines_pct IS NOT NULL), 0)
		  END AS ai_percent
		FROM commits c
		JOIN team_members tm ON tm.team_id = c.team_id AND tm.user_id = $1
		WHERE c.sha = ANY($2)
	`, userID, shas)

	var agg Aggregate
	var aiPct *float64
	if err := row.Scan(&agg.TotalCommits, &agg.WithAttribution, &agg.LinesAdded, &agg.LinesRemoved, &aiPct); err != nil {
		return Aggregate{}, err
	}
	agg.AIPercentWeighted = aiPct
	return agg, nil
}
