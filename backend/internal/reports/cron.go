package reports

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/reports/reportsapp"
	"github.com/eye-of-providence/backend/internal/store"
)

type Cron struct {
	Interval   time.Duration
	Store      ReportStore
	EventStore store.EventStore
	Gemini     *GeminiClient
	Logger     *zap.Logger
}

func (c *Cron) Run(ctx context.Context) {
	if c.Interval <= 0 {
		c.Interval = 6 * time.Hour
	}
	appSvc := NewReportsApp(c.Store, c.EventStore, c.Gemini)
	t := time.NewTicker(c.Interval)
	defer t.Stop()

	c.tick(ctx, appSvc)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.tick(ctx, appSvc)
		}
	}
}

func (c *Cron) tick(ctx context.Context, appSvc *reportsapp.Service) {
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	users, err := appSvc.ActiveUserIDs(ctx, since)
	if err != nil {
		c.Logger.Warn("cron: ActiveUserIDs failed", zap.Error(err))
		return
	}

	from, to, key := weeklyPeriod(time.Now().UTC())
	now := time.Now().UTC()

	for _, uid := range users {
		if appSvc.HasPeriod(uid, key) {
			continue
		}
		if err := appSvc.GenerateWeeklyIfMissing(ctx, uid, from, to, key, now); err != nil {
			c.Logger.Warn("cron: generate failed", zap.String("user", uid), zap.Error(err))
		} else {
			c.Logger.Info("cron: generated", zap.String("user", uid), zap.String("period", key))
		}
	}
}

func weeklyPeriod(now time.Time) (time.Time, time.Time, string) {
	year, week := now.ISOWeek()
	offset := int(now.Weekday())
	if offset == 0 {
		offset = 7
	}
	monday := now.AddDate(0, 0, 1-offset)
	from := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 7).Add(-time.Second)
	return from, to, fmt.Sprintf("weekly_%04d_W%02d", year, week)
}
