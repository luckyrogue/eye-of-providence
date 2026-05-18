package analyticsapp

import (
	"context"
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/analytics/domain"
)

type fakeRead struct{}

func (fakeRead) ListRecent(context.Context, string, int) ([]domain.Event, error) {
	return []domain.Event{{UserID: "u1"}}, nil
}

func (fakeRead) LanguageBreakdown(context.Context, string, time.Time) ([]domain.LangCell, error) {
	return []domain.LangCell{{Lang: "go"}}, nil
}

func (fakeRead) DailyTrend(context.Context, string, time.Time, string) ([]domain.TrendPoint, error) {
	return []domain.TrendPoint{{Date: "2026-01-01"}}, nil
}

func (fakeRead) Heatmap(context.Context, string, time.Time, string) ([]domain.HeatmapCell, error) {
	return []domain.HeatmapCell{{Hour: 9}}, nil
}

func (fakeRead) AggregateByCategory(context.Context, string, time.Time) (map[string]uint64, error) {
	return map[string]uint64{"coding": 1}, nil
}

func TestListRecent(t *testing.T) {
	svc := New(Deps{Events: fakeRead{}})
	events, err := svc.ListRecent(context.Background(), "u1", 10)
	if err != nil || len(events) != 1 {
		t.Fatalf("events=%v err=%v", events, err)
	}
}

func TestCategories(t *testing.T) {
	svc := New(Deps{Events: fakeRead{}})
	agg, err := svc.Categories(context.Background(), "u1", WindowFromDays(7))
	if err != nil || agg["coding"] != 1 {
		t.Fatalf("agg=%v err=%v", agg, err)
	}
}
