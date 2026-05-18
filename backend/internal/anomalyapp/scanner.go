package anomalyapp

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

)

const EventName = "anomaly.detected"

type Scanner struct {
	Interval   time.Duration
	Events     TrendSource
	Webhooks   WebhookDispatcher
	Push       PushSender
	Logger     *zap.Logger

	seen map[string]bool
}

func (s *Scanner) Run(ctx context.Context) {
	if s.Interval <= 0 {
		s.Interval = 6 * time.Hour
	}
	if s.seen == nil {
		s.seen = make(map[string]bool)
	}
	t := time.NewTicker(s.Interval)
	defer t.Stop()

	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.tick(ctx)
		}
	}
}

func (s *Scanner) tick(ctx context.Context) {
	if s.Webhooks == nil || s.Events == nil {
		return
	}
	since := time.Now().UTC().Add(-15 * 24 * time.Hour)
	users, err := s.Events.ActiveUserIDs(ctx, since)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Warn("anomaly scan: ActiveUserIDs failed", zap.Error(err))
		}
		return
	}

	now := time.Now().UTC()

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	var seenMu sync.Mutex

	for _, uidStr := range users {
		g.Go(func() error {
			uid, err := uuid.Parse(uidStr)
			if err != nil {
				return nil
			}
			points, err := s.Events.DailyTrend(gctx, uidStr, since, "UTC")
			if err != nil {
				if s.Logger != nil {
					s.Logger.Warn("anomaly scan: DailyTrend failed", zap.String("user", uidStr), zap.Error(err))
				}
				return nil
			}
			anomalies := DetectTrends(points, now)
			for _, a := range anomalies {
				key := uidStr + "|" + a.Date + "|" + string(a.Kind)
				seenMu.Lock()
				dup := s.seen[key]
				if !dup {
					s.seen[key] = true
				}
				seenMu.Unlock()
				if dup {
					continue
				}
				s.Webhooks.Dispatch(uid, EventName, a)
				if s.Push != nil {
					s.Push.SendToUser(uid, PushPayload(a))
				}
				if s.Logger != nil {
					s.Logger.Info("anomaly fired",
						zap.String("user", uidStr),
						zap.String("kind", string(a.Kind)),
						zap.Float64("z", a.ZScore))
				}
			}
			return nil
		})
	}
	_ = g.Wait()

	cutoff := now.AddDate(0, 0, -7).Format("2006-01-02")
	for key := range s.seen {
		parts := splitN(key, "|", 3)
		if len(parts) >= 2 && parts[1] < cutoff {
			delete(s.seen, key)
		}
	}
}

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
