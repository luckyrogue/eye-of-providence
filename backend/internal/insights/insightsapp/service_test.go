package insightsapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/insights/domain"
	"github.com/eye-of-providence/backend/internal/insights/insightsapp"
)

type fakeEvents struct{}

func (fakeEvents) AggregateByCategory(context.Context, string, time.Time) (map[string]uint64, error) {
	return map[string]uint64{"ai": 1}, nil
}

func (fakeEvents) LanguageBreakdown(context.Context, string, time.Time) ([]domain.LangCell, error) {
	return nil, nil
}

func (fakeEvents) DailyTrend(context.Context, string, time.Time, string) ([]domain.TrendPoint, error) {
	return nil, nil
}

type fakeRanges struct{}

func (fakeRanges) AggregateWindow(context.Context, string, time.Time, time.Time) (map[string]uint64, error) {
	return map[string]uint64{"ai": 1}, nil
}

func TestGenerate(t *testing.T) {
	svc := insightsapp.New(insightsapp.Deps{Events: fakeEvents{}, Ranges: fakeRanges{}})
	out, err := svc.Generate(context.Background(), "u", "UTC", time.Now().UTC())
	if err != nil || len(out) == 0 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
