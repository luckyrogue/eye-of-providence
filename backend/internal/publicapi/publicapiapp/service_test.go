package publicapiapp

import (
	"context"
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/publicapi/domain"
)

type fakeRead struct{}

func (fakeRead) ListRecent(context.Context, string, int) ([]domain.Event, error) {
	return []domain.Event{{UserID: "u"}}, nil
}

func (fakeRead) AggregateByCategory(context.Context, string, time.Time) (map[string]uint64, error) {
	return map[string]uint64{"ai": 2}, nil
}

func (fakeRead) LanguageBreakdown(context.Context, string, time.Time) ([]domain.LangCell, error) {
	return []domain.LangCell{{Lang: "ts"}}, nil
}

func (fakeRead) DailyTrend(context.Context, string, time.Time, string) ([]domain.TrendPoint, error) {
	return []domain.TrendPoint{{Date: "d"}}, nil
}

func TestListRecent(t *testing.T) {
	svc := New(Deps{Events: fakeRead{}})
	events, err := svc.ListRecent(context.Background(), "u", 10)
	if err != nil || len(events) != 1 {
		t.Fatalf("events=%v err=%v", events, err)
	}
}
