package anomaly

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/store"
)

// Dispatcher — webhooks-style fire-and-forget sender. internal/webhooks.Service
// этому соответствует.
type Dispatcher interface {
	Dispatch(userID uuid.UUID, event string, payload any)
}

// EventName — webhook event'а. Должен быть зарегистрирован в
// internal/webhooks/webhooks.go validEvents для доставки.
const EventName = "anomaly.detected"

// Cron — daily anomaly checker. Раз в Interval тикает: для каждого active
// user'а fetch'ит DailyTrend(15 days), детектит, отправляет через Dispatcher.
//
// Граничит с reports/cron.go (тот же активность-cycle), но live в отдельном
// пакете чтобы избежать import cycle и cleanly tested.
type Cron struct {
	Interval   time.Duration
	EventStore store.EventStore
	Webhooks   Dispatcher
	Logger     *zap.Logger
	// dedup: userID+anomaly.Date+anomaly.Kind → не шлём одну и ту же аномалию
	// дважды если cron tick'ает несколько раз в день.
	seen map[string]bool
}

func (c *Cron) Run(ctx context.Context) {
	if c.Interval <= 0 {
		c.Interval = 6 * time.Hour
	}
	if c.seen == nil {
		c.seen = make(map[string]bool)
	}
	t := time.NewTicker(c.Interval)
	defer t.Stop()

	c.tick(ctx)
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
	if c.Webhooks == nil {
		return // в-memory mode — webhooks не настроены
	}
	since := time.Now().UTC().Add(-15 * 24 * time.Hour)
	users, err := c.EventStore.ActiveUserIDs(ctx, since)
	if err != nil {
		c.Logger.Warn("anomaly cron: ActiveUserIDs failed", zap.Error(err))
		return
	}

	now := time.Now().UTC()
	for _, uidStr := range users {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			continue
		}
		points, err := c.EventStore.DailyTrend(ctx, uidStr, since, "UTC")
		if err != nil {
			c.Logger.Warn("anomaly cron: DailyTrend failed", zap.String("user", uidStr), zap.Error(err))
			continue
		}
		trends := make([]Trend, 0, len(points))
		for _, p := range points {
			trends = append(trends, Trend{Date: p.Date, Category: p.Category, MS: p.MS})
		}
		anomalies := Detect(MakeInputs(trends, now))
		for _, a := range anomalies {
			key := uidStr + "|" + a.Date + "|" + string(a.Kind)
			if c.seen[key] {
				continue
			}
			c.seen[key] = true
			c.Webhooks.Dispatch(uid, EventName, a)
			c.Logger.Info("anomaly fired",
				zap.String("user", uidStr),
				zap.String("kind", string(a.Kind)),
				zap.Float64("z", a.ZScore))
		}
	}

	// GC seen-map: удаляем entries старше 7 дней (по date в key).
	cutoff := now.AddDate(0, 0, -7).Format("2006-01-02")
	for key := range c.seen {
		// key format: userID|date|kind. Извлекаем date через split.
		parts := splitN(key, "|", 3)
		if len(parts) >= 2 && parts[1] < cutoff {
			delete(c.seen, key)
		}
	}
}

// splitN — лёгкий split без import strings, чтобы tests могли быть
// minimal-dep. (Мы используем strings везде в проекте, но внутри cron
// hot-path хочется maintain'ить cheap.) Actually это дешевле чем strings.
func splitN(s, sep string, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		idx := -1
		for j := 0; j < len(s); j++ {
			if j+len(sep) <= len(s) && s[j:j+len(sep)] == sep {
				idx = j
				break
			}
		}
		if idx < 0 {
			break
		}
		out = append(out, s[:idx])
		s = s[idx+len(sep):]
	}
	out = append(out, s)
	return out
}
