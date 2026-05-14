package prcomment

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Aggregate struct {
	TotalCommits    int `json:"total_commits"`
	WithAttribution int `json:"with_attribution"`
	LinesAdded      int `json:"lines_added"`
	LinesRemoved    int `json:"lines_removed"`

	AIPercentWeighted *float64 `json:"ai_percent,omitempty"`
}

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
