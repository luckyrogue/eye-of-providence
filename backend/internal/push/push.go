// Package push — Web Push API delivery (PWA notifications).
//
// VAPID keys (ECDSA P-256) генерятся раз через `cmd/vapid-gen` и хранятся в
// env. Public key экспортируется на фронт через GET /v1/me/push/vapid-key
// для PushManager.subscribe(). Server подписывает каждый push JWT'ом
// от private key.
//
// Storage: push_subscriptions per (user, endpoint). Delete-on-410 (subscription
// expired) — стандартный сигнал от push service когда юзер отозвал permission.
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

// Subscription — DB row для UI listing (без auth/p256dh — секреты для send'а).
type Subscription struct {
	ID         uuid.UUID `json:"id"`
	Endpoint   string    `json:"endpoint"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  string    `json:"created_at"`
	LastUsedAt *string   `json:"last_used_at,omitempty"`
}

// Service — push delivery + DB ops. VAPID keys load'ятся из cfg.
type Service struct {
	Pool            *pgxpool.Pool
	Logger          *zap.Logger
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string // "mailto:admin@eop.example.com" — required by RFC 8292
	HTTPClient      *http.Client
	// Hardening (Phase prod): bounded goroutine pool + shutdown ctx.
	// SendToUser fire-and-forget теперь tracked, main.go может Wait'нуть
	// на graceful shutdown.
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
	inflight       sync.WaitGroup
	sendSem        chan struct{}
}

const maxConcurrentSend = 64

// Init — должен быть вызван caller'ом ДО первого SendToUser. Создаёт
// shutdown context + semaphore. Если не вызвать, SendToUser fall back на
// background ctx без bounded pool (legacy behavior) — для in-memory tests.
func (s *Service) Init() {
	if s.shutdownCtx != nil {
		return // idempotent
	}
	s.shutdownCtx, s.shutdownCancel = context.WithCancel(context.Background())
	s.sendSem = make(chan struct{}, maxConcurrentSend)
}

// Shutdown — graceful: ждёт inflight delivery до ctx.Done() или sub-deadline.
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

// Payload — отправляется на subscription. JSON, доходит до sw.js push event.
type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	URL   string `json:"url,omitempty"` // куда клик ведёт (default /dashboard)
	Tag   string `json:"tag,omitempty"` // dedupe по tag, например "anomaly.ai_high"
}

// Subscribe — store new subscription. Идемпотентно по endpoint (UNIQUE
// constraint); если уже существует, обновляем p256dh/auth + last_used.
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

// Unsubscribe — удаляет по endpoint. Возвращает true если запись существовала.
func (s *Service) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) (bool, error) {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`,
		userID, endpoint)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// List — все subscriptions пользователя для UI listing.
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

// SendToUser — fire-and-forget delivery всем subscriptions юзера. payload
// принимает any (Payload struct, map[string]any с keys title/body/url/tag,
// или любая JSON-marshallable структура) — так satisfy'ятся interface'ы
// caller'ов (anomaly.PushSender, etc.) без import cycle.
//
// Hardening: tracked в inflight WaitGroup + bounded sendSem (64 concurrent).
// На shutdown service ждёт delivery до deadline'а. Если Init() не вызван
// (in-memory test mode) — fallback на legacy fire-and-forget.
func (s *Service) SendToUser(userID uuid.UUID, payload any) {
	if s.sendSem == nil {
		// Legacy path для тестов / in-memory режима.
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
		// 60s per-send timeout, derived от shutdownCtx чтобы cancel прерывал
		// retry-loops в webpush.SendNotification.
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
	// Fan-out: per-device push в parallel. Limit=4 — typical user 1-3
	// devices, но чтобы избежать spawn'а N горутин если кто-то залип на
	// 10+ devices.
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

// deliver — single push. Обрабатывает 404/410 как "subscription gone" — удаляем.
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
		TTL:             86400, // 24 часа — push service может отложить доставку
		Urgency:         webpush.UrgencyNormal,
	})
	if err != nil {
		s.Logger.Warn("push: send failed", zap.String("endpoint", short(endpoint)), zap.Error(err))
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 404 || res.StatusCode == 410 {
		// Subscription уволенa юзером или expired. Удаляем чтобы не спамить
		// dead endpoint каждый push.
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

// short — log-friendly shortening endpoint (FCM URLs длинные, в логи попадает
// только host + первые chars path'а).
func short(endpoint string) string {
	if len(endpoint) > 60 {
		return endpoint[:60] + "…"
	}
	return endpoint
}
