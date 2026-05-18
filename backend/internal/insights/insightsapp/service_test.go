package insightsapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/insights/domain"
	"github.com/eye-of-providence/backend/internal/insights/insightsapp"
)

// 1 час в ms — выше всех порогов (totalActivity 0.5h, aiRatio 30 min,
// aiTrend 5 min). Если fake вернёт `1` ms — все insights дают nil и
// результат пустой, тест падает.
const fakeHourMS uint64 = 60 * 60 * 1000

type fakeEvents struct{}

func (fakeEvents) AggregateByCategory(context.Context, string, time.Time) (map[string]uint64, error) {
	return map[string]uint64{"ai": fakeHourMS}, nil
}

func (fakeEvents) LanguageBreakdown(context.Context, string, time.Time) ([]domain.LangCell, error) {
	return nil, nil
}

func (fakeEvents) DailyTrend(context.Context, string, time.Time, string) ([]domain.TrendPoint, error) {
	return nil, nil
}

type fakeRanges struct{}

func (fakeRanges) AggregateWindow(context.Context, string, time.Time, time.Time) (map[string]uint64, error) {
	return map[string]uint64{"ai": fakeHourMS}, nil
}

func TestGenerate(t *testing.T) {
	svc := insightsapp.New(insightsapp.Deps{Events: fakeEvents{}, Ranges: fakeRanges{}})
	out, err := svc.Generate(context.Background(), "u", "UTC", time.Now().UTC())
	if err != nil || len(out) == 0 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
