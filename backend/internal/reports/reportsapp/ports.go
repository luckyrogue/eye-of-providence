package reportsapp

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/reports/domain"
)

type ReportRepository interface {
	Save(r domain.Report)
	ListForUser(userID string, limit int) []domain.Report
	Get(id, userID string) (domain.Report, bool)
}

type ReportGenerator interface {
	Generate(ctx context.Context, nc *domain.NumericContext) (string, error)
	Model() string
}

type ContextBuilder interface {
	Build(ctx context.Context, userID, period string, from, to time.Time) (*domain.NumericContext, error)
}

type ActiveUsers interface {
	ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error)
}
