package webhooksapp

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Dispatcher struct {
	Lister    ActiveWebhookLister
	Format    PayloadFormatter
	Sign      Signer
	HTTP      HTTPPoster
	Record    DeliveryRecorder
	Logger    *zap.Logger
	ValidEvent func(event string) bool
}

func (d *Dispatcher) DispatchSync(ctx context.Context, userID uuid.UUID, event string, payload any) {
	if d.ValidEvent != nil && !d.ValidEvent(event) {
		if d.Logger != nil {
			d.Logger.Warn("unknown event for dispatch", zap.String("event", event))
		}
		return
	}
	if d.Lister == nil {
		return
	}
	targets, err := d.Lister.ListActive(ctx, userID, event)
	if err != nil {
		if d.Logger != nil {
			d.Logger.Error("webhook lookup failed", zap.Error(err))
		}
		return
	}
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for _, t := range targets {
		t := t
		g.Go(func() error {
			d.deliver(gctx, t, event, payload)
			return nil
		})
	}
	_ = g.Wait()
}

func (d *Dispatcher) deliver(ctx context.Context, t DeliveryTarget, event string, payload any) {
	if d.Format == nil || d.HTTP == nil || d.Sign == nil {
		return
	}
	body, err := d.Format.Format(t.Format, event, payload)
	if err != nil {
		if d.Logger != nil {
			d.Logger.Error("format payload", zap.String("format", t.Format), zap.Error(err))
		}
		return
	}
	sig := d.Sign.Sign(t.Secret, body)
	delays := []time.Duration{0, time.Second, 3 * time.Second, 9 * time.Second}
	lastStatus := 0
	for attempt, delay := range delays {
		if delay > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}
		status, err := d.HTTP.Post(ctx, t.URL, map[string]string{
			"Content-Type":     "application/json",
			"X-EoP-Signature":  "sha256=" + sig,
			"X-EoP-Event":      event,
			"User-Agent":       "eop-webhooks/1.0",
		}, body)
		lastStatus = status
		if err == nil && status < 500 {
			break
		}
		if err != nil {
			lastStatus = -1
		}
		if d.Logger != nil {
			d.Logger.Debug("webhook retry",
				zap.String("url", t.URL), zap.Int("status", status), zap.Int("attempt", attempt), zap.Error(err))
		}
	}
	if d.Record != nil {
		_ = d.Record.Record(ctx, t.ID, lastStatus)
	}
}
