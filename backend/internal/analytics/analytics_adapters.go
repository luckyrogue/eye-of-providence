package analytics

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/analytics/analyticsapp"
	"github.com/eye-of-providence/backend/internal/analytics/domain"
	"github.com/eye-of-providence/backend/internal/store"
)

type eventReadAdapter struct{ st store.EventStore }

func (a eventReadAdapter) ListRecent(ctx context.Context, userID string, limit int) ([]domain.Event, error) {
	rows, err := a.st.ListRecent(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Event, len(rows))
	for i := range rows {
		out[i] = mapEvent(rows[i])
	}
	return out, nil
}

func (a eventReadAdapter) LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]domain.LangCell, error) {
	cells, err := a.st.LanguageBreakdown(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	out := make([]domain.LangCell, len(cells))
	for i := range cells {
		out[i] = domain.LangCell(cells[i])
	}
	return out, nil
}

func (a eventReadAdapter) DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]domain.TrendPoint, error) {
	points, err := a.st.DailyTrend(ctx, userID, since, tz)
	if err != nil {
		return nil, err
	}
	out := make([]domain.TrendPoint, len(points))
	for i := range points {
		out[i] = domain.TrendPoint(points[i])
	}
	return out, nil
}

func (a eventReadAdapter) Heatmap(ctx context.Context, userID string, since time.Time, tz string) ([]domain.HeatmapCell, error) {
	cells, err := a.st.Heatmap(ctx, userID, since, tz)
	if err != nil {
		return nil, err
	}
	out := make([]domain.HeatmapCell, len(cells))
	for i := range cells {
		out[i] = domain.HeatmapCell(cells[i])
	}
	return out, nil
}

func (a eventReadAdapter) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	return a.st.AggregateByCategory(ctx, userID, since)
}

func mapEvent(e store.Event) domain.Event {
	return domain.Event{
		TS:           e.TS,
		UserID:       e.UserID,
		DeviceID:     e.DeviceID,
		SessionID:    e.SessionID,
		AppBundle:    e.AppBundle,
		Category:     e.Category,
		Source:       e.Source,
		AIProvider:   e.AIProvider,
		AIChannel:    e.AIChannel,
		ProjectID:    e.ProjectID,
		FileLang:     e.FileLang,
		DurationMS:   e.DurationMS,
		CharsIn:      e.CharsIn,
		LinesAdded:   e.LinesAdded,
		LinesRemoved: e.LinesRemoved,
		Meta:         e.Meta,
	}
}

func newAnalyticsApp(st store.EventStore) *analyticsapp.Service {
	return analyticsapp.New(analyticsapp.Deps{Events: eventReadAdapter{st: st}})
}
