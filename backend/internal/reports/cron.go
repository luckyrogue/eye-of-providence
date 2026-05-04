package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/store"
)

// Cron — фоновый генератор weekly отчётов.
// Раз в `interval` тикает: для каждого активного user (за последние 7 дней)
// проверяет, есть ли отчёт за текущую ISO week, если нет — генерирует.
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
	t := time.NewTicker(c.Interval)
	defer t.Stop()

	c.tick(ctx) // первый прогон на старте
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.tick(ctx)
		}
	}
}

func (c *Cron) tick(ctx context.Context) {
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	users, err := c.EventStore.ActiveUserIDs(ctx, since)
	if err != nil {
		c.Logger.Warn("cron: ActiveUserIDs failed", zap.Error(err))
		return
	}

	from, to, key := weeklyPeriod(time.Now().UTC())

	for _, uid := range users {
		if c.userHasReport(uid, key) {
			continue
		}
		if err := c.generateFor(ctx, uid, key, from, to); err != nil {
			c.Logger.Warn("cron: generate failed", zap.String("user", uid), zap.Error(err))
		} else {
			c.Logger.Info("cron: generated", zap.String("user", uid), zap.String("period", key))
		}
	}
}

func (c *Cron) userHasReport(userID, period string) bool {
	for _, r := range c.Store.ListForUser(userID, 50) {
		if r.Period == period {
			return true
		}
	}
	return false
}

func (c *Cron) generateFor(ctx context.Context, userID, period string, from, to time.Time) error {
	nc, err := BuildContext(ctx, c.EventStore, userID, period, from, to)
	if err != nil {
		return err
	}
	body, err := c.Gemini.Generate(ctx, nc)
	if err != nil {
		return err
	}
	c.Store.Save(Report{
		ID:            uuid.NewString(),
		UserID:        userID,
		Period:        period,
		Model:         c.Gemini.Model,
		BodyMD:        body,
		GeneratedAt:   time.Now().UTC(),
		PromptVersion: promptVersion,
	})
	return nil
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

