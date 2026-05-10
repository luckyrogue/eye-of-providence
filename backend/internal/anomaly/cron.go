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

// PushSender — Web Push delivery (PWA). Optional, nil → push не шлётся.
// Push дублирует Slack для retention loop'а (юзер видит OS-уровневое
// уведомление без открытия Slack).
type PushSender interface {
	SendToUser(userID uuid.UUID, payload any)
}

// EventName — webhook event'а. Должен быть зарегистрирован в
// internal/webhooks/webhooks.go validEvents для доставки.
const EventName = "anomaly.detected"

// Cron — daily anomaly checker. Раз в Interval тикает: для каждого active
// user'а fetch'ит DailyTrend(15 days), детектит, отправляет через Dispatcher
// + опционально через PushSender.
type Cron struct {
	Interval   time.Duration
	EventStore store.EventStore
	Webhooks   Dispatcher
	Push       PushSender // optional
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
			if c.Push != nil {
				c.Push.SendToUser(uid, pushPayloadFor(a))
			}
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

// pushPayloadFor — формирует push.Payload (но не импортируем тот pkg —
// возвращаем generic map чтобы избежать import cycle anomaly → push). Service
// SendToUser принимает any → JSON-marshal.
func pushPayloadFor(a Anomaly) any {
	titles := map[Kind]string{
		KindAIHigh:       "AI usage spike",
		KindAILow:        "AI usage dip",
		KindManualHigh:   "Manual coding spike",
		KindManualLow:    "Manual coding dip",
		KindRefactorHigh: "Refactoring day",
		KindActivityHigh: "Productivity spike",
		KindActivityLow:  "Activity dropped",
	}
	title, ok := titles[a.Kind]
	if !ok {
		title = "Coding anomaly"
	}
	yHrs := float64(a.YesterdayMS) / 3600000
	bHrs := float64(a.BaselineMS) / 3600000
	return map[string]any{
		"title": title,
		"body":  fmtAnomalyBody(yHrs, bHrs, a.Category),
		"url":   "/dashboard",
		"tag":   "anomaly." + string(a.Kind), // dedupe push'ей одного типа
	}
}

func fmtAnomalyBody(yHrs, bHrs float64, category string) string {
	// "AI 5.2h vs 1.1h baseline" — короткий single-line под Lock-screen.
	return formatHours(yHrs) + " vs " + formatHours(bHrs) + " baseline (" + category + ")"
}

func formatHours(h float64) string {
	if h < 1 {
		min := int(h*60 + 0.5)
		return itoa(min) + "min"
	}
	tenths := int(h*10 + 0.5)
	whole := tenths / 10
	frac := tenths % 10
	if frac == 0 {
		return itoa(whole) + "h"
	}
	return itoa(whole) + "." + itoa(frac) + "h"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
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
