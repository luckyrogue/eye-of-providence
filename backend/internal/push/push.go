package push

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Subscription struct {
	ID         uuid.UUID `json:"id"`
	Endpoint   string    `json:"endpoint"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  string    `json:"created_at"`
	LastUsedAt *string   `json:"last_used_at,omitempty"`
}

type Service struct {
	Pool            *pgxpool.Pool
	Logger          *zap.Logger
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
	HTTPClient      *http.Client

	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
	inflight       sync.WaitGroup
	sendSem        chan struct{}
}

const maxConcurrentSend = 64

func (s *Service) Init() {
	if s.shutdownCtx != nil {
		return
	}
	s.shutdownCtx, s.shutdownCancel = context.WithCancel(context.Background())
	s.sendSem = make(chan struct{}, maxConcurrentSend)
}

func (s *Service) Shutdown(ctx context.Context) error {
	if s.shutdownCancel == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		s.inflight.Wait()
		close(done)
	}()
	select {
	case <-done:
		s.shutdownCancel()
		return nil
	case <-ctx.Done():
		s.shutdownCancel()
		return ctx.Err()
	}
}

type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	URL   string `json:"url,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth, userAgent string) error {
	if endpoint == "" || p256dh == "" || auth == "" {
		return errors.New("endpoint/p256dh/auth required")
	}
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (endpoint) DO UPDATE
		SET p256dh = EXCLUDED.p256dh,
		    auth = EXCLUDED.auth,
		    user_agent = EXCLUDED.user_agent,
		    last_used_at = now()`,
		userID, endpoint, p256dh, auth, userAgent)
	return err
}

func (s *Service) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) (bool, error) {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`,
		userID, endpoint)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, endpoint, COALESCE(user_agent, ''),
		       to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       to_char(last_used_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM push_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Subscription{}
	for rows.Next() {
		var sub Subscription
		var lastUsed *string
		if err := rows.Scan(&sub.ID, &sub.Endpoint, &sub.UserAgent, &sub.CreatedAt, &lastUsed); err != nil {
			return nil, err
		}
		sub.LastUsedAt = lastUsed
		out = append(out, sub)
	}
	return out, nil
}

func (s *Service) SendToUser(userID uuid.UUID, payload any) {
	if s.sendSem == nil {

		go s.sendSync(context.Background(), userID, payload)
		return
	}
	select {
	case s.sendSem <- struct{}{}:
	case <-s.shutdownCtx.Done():
		return
	}
	s.inflight.Add(1)
	go func() {
		defer func() {
			<-s.sendSem
			s.inflight.Done()
		}()

		ctx, cancel := context.WithTimeout(s.shutdownCtx, 60*time.Second)
		defer cancel()
		s.sendSync(ctx, userID, payload)
	}()
}

func (s *Service) sendSync(ctx context.Context, userID uuid.UUID, payload any) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, endpoint, p256dh, auth FROM push_subscriptions WHERE user_id = $1`, userID)
	if err != nil {
		s.Logger.Warn("push: lookup failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type target struct {
		id       uuid.UUID
		endpoint string
		p256dh   string
		auth     string
	}
	targets := []target{}
	for rows.Next() {
		var t target
		if err := rows.Scan(&t.id, &t.endpoint, &t.p256dh, &t.auth); err == nil {
			targets = append(targets, t)
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	for _, t := range targets {
		g.Go(func() error {
			s.deliver(gctx, t.id, t.endpoint, t.p256dh, t.auth, body)
			return nil
		})
	}
	_ = g.Wait()
}

func (s *Service) deliver(ctx context.Context, id uuid.UUID, endpoint, p256dh, auth string, body []byte) {
	sub := &webpush.Subscription{
		Endpoint: endpoint,
		Keys:     webpush.Keys{P256dh: p256dh, Auth: auth},
	}
	res, err := webpush.SendNotificationWithContext(ctx, body, sub, &webpush.Options{
		HTTPClient:      s.HTTPClient,
		Subscriber:      s.VAPIDSubject,
		VAPIDPublicKey:  s.VAPIDPublicKey,
		VAPIDPrivateKey: s.VAPIDPrivateKey,
		TTL:             86400,
		Urgency:         webpush.UrgencyNormal,
	})
	if err != nil {
		s.Logger.Warn("push: send failed", zap.String("endpoint", short(endpoint)), zap.Error(err))
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 404 || res.StatusCode == 410 {

		_, _ = s.Pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE id = $1`, id)
		s.Logger.Info("push: subscription expired, removed", zap.String("endpoint", short(endpoint)))
		return
	}
	if res.StatusCode >= 400 {
		s.Logger.Warn("push: provider error",
			zap.Int("status", res.StatusCode),
			zap.String("endpoint", short(endpoint)))
		return
	}
	_, _ = s.Pool.Exec(ctx, `UPDATE push_subscriptions SET last_used_at = now() WHERE id = $1`, id)
}

func short(endpoint string) string {
	if len(endpoint) > 60 {
		return endpoint[:60] + "…"
	}
	return endpoint
}
