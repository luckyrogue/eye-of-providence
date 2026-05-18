package reports

import (
	"context"
	"errors"
	"time"

	"github.com/eye-of-providence/backend/internal/reports/domain"
	"github.com/eye-of-providence/backend/internal/reports/reportsapp"
	"github.com/eye-of-providence/backend/internal/store"
)

type reportRepoAdapter struct{ s ReportStore }

func (a reportRepoAdapter) Save(r domain.Report) {
	a.s.Save(Report(r))
}

func (a reportRepoAdapter) ListForUser(userID string, limit int) []domain.Report {
	rows := a.s.ListForUser(userID, limit)
	out := make([]domain.Report, len(rows))
	for i := range rows {
		out[i] = domain.Report(rows[i])
	}
	return out
}

func (a reportRepoAdapter) Get(id, userID string) (domain.Report, bool) {
	r, ok := a.s.Get(id, userID)
	return domain.Report(r), ok
}

type contextBuilderAdapter struct{ st store.EventStore }

func (a contextBuilderAdapter) Build(ctx context.Context, userID, period string, from, to time.Time) (*domain.NumericContext, error) {
	return BuildContext(ctx, a.st, userID, period, from, to)
}

type reportGeneratorAdapter struct{ g *GeminiClient }

func (a reportGeneratorAdapter) Generate(ctx context.Context, nc *domain.NumericContext) (string, error) {
	body, err := a.g.Generate(ctx, nc)
	if err != nil {
		return "", errors.Join(errReportGeneration, err)
	}
	return body, nil
}

func (a reportGeneratorAdapter) Model() string { return a.g.Model }

type activeUsersAdapter struct{ st store.EventStore }

func (a activeUsersAdapter) ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error) {
	return a.st.ActiveUserIDs(ctx, since)
}

func NewReportsApp(store ReportStore, events store.EventStore, gemini *GeminiClient) *reportsapp.Service {
	return reportsapp.New(reportsapp.Deps{
		Repo:      reportRepoAdapter{s: store},
		Builder:   contextBuilderAdapter{st: events},
		Generator: reportGeneratorAdapter{g: gemini},
		Users:     activeUsersAdapter{st: events},
	})
}
