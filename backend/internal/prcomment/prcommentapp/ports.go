package prcommentapp

import (
	"context"

	"github.com/google/uuid"
)

type Aggregate struct {
	TotalCommits      int      `json:"total_commits"`
	WithAttribution   int      `json:"with_attribution"`
	LinesAdded        int      `json:"lines_added"`
	LinesRemoved      int      `json:"lines_removed"`
	AIPercentWeighted *float64 `json:"ai_percent,omitempty"`
}

type Formatter interface {
	Markdown(ctx context.Context, userID uuid.UUID, shas []string) (comment string, agg Aggregate, err error)
}

type Poster interface {
	Post(ctx context.Context, provider, host, repo string, prNumber int, token, body string) error
}
