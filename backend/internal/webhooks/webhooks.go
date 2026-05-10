// Package webhooks — outbound event delivery (commit.ingested, report.generated).
//
// Caller (commits/reports handlers) дёргает Dispatch(ctx, userID, event, payload)
// после успешного persist. Dispatch неблокирующий — спавнит горутину которая
// 1) lookup active webhooks для user'а
// 2) POST'ит JSON с HMAC-SHA256 signature header
// 3) retry 3 раза с backoff 1s/3s/9s на 5xx/network errors
// 4) update last_delivery_at / last_status в БД
//
// HMAC: header "X-EoP-Signature: sha256=<hex>", computed over raw body. Receiver
// должен validate constant-time. Secret генерится при создании webhook'а.
package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Известные event types. Constant'ы т.к. они часть public contract.
const (
	EventCommitIngested  = "commit.ingested"
	EventReportGenerated = "report.generated"
	EventAnomalyDetected = "anomaly.detected"
)

var validEvents = map[string]bool{
	EventCommitIngested:  true,
	EventReportGenerated: true,
	EventAnomalyDetected: true,
}

// Webhook — DB row, без plaintext secret'а (получается ровно раз при create).
type Webhook struct {
	ID             uuid.UUID  `json:"id"`
	URL            string     `json:"url"`
	Events         []string   `json:"events"`
	Format         string     `json:"format"` // "raw" | "slack"
	Active         bool       `json:"active"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	LastStatus     *int       `json:"last_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Service — DB pool + dispatcher config. Несёт state для in-flight retries.
type Service struct {
	Pool       *pgxpool.Pool
	Logger     *zap.Logger
	HTTPClient *http.Client // если nil → дефолт с timeout 5s
}

// New — конструктор с дефолтным HTTP client.
func New(pool *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{
		Pool:       pool,
		Logger:     logger,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// CreateWebhook — INSERT с генерёным secret'ом. Возвращает (plaintext_secret, row).
// Plaintext secret отдаётся ровно один раз (caller должен показать в UI).
// format пустой → "raw".
func (s *Service) Create(ctx context.Context, userID uuid.UUID, url string, events []string, format string) (string, Webhook, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return "", Webhook{}, errors.New("url must be http(s)://")
	}
	if len(url) > 2048 {
		return "", Webhook{}, errors.New("url too long")
	}
	for _, e := range events {
		if !validEvents[e] {
			return "", Webhook{}, fmt.Errorf("unknown event: %s", e)
		}
	}
	if len(events) == 0 {
		return "", Webhook{}, errors.New("at least one event required")
	}
	if format == "" {
		format = string(FormatRaw)
	}
	if !validFormat(format) {
		return "", Webhook{}, fmt.Errorf("unknown format: %s", format)
	}

	secret, err := generateSecret()
	if err != nil {
		return "", Webhook{}, err
	}

	out := Webhook{
		ID:     uuid.New(),
		URL:    url,
		Events: events,
		Format: format,
		Active: true,
	}
	err = s.Pool.QueryRow(ctx, `
		INSERT INTO webhooks (id, user_id, url, secret, events, format)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`,
		out.ID, userID, url, secret, events, format,
	).Scan(&out.CreatedAt)
	if err != nil {
		return "", Webhook{}, err
	}
	return secret, out, nil
}

// List — active webhooks пользователя.
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Webhook, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, url, events, format, active, last_delivery_at, last_status, created_at
		FROM webhooks WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Webhook{}
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(&w.ID, &w.URL, &w.Events, &w.Format, &w.Active, &w.LastDeliveryAt, &w.LastStatus, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

// Delete — hard delete (не soft, в отличие от api_tokens — webhook secret уже
// в plaintext БД, soft-delete не даёт security gain).
func (s *Service) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM webhooks WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// Dispatch — fire-and-forget delivery всех matching webhook'ов. Spawns горутину;
// caller не ждёт. Errors логируются, не возвращаются. Используется в hot path
// (commit ingest, report gen) где блокировка недопустима.
func (s *Service) Dispatch(userID uuid.UUID, event string, payload any) {
	if !validEvents[event] {
		s.Logger.Warn("unknown event for dispatch", zap.String("event", event))
		return
	}
	go s.dispatchSync(userID, event, payload)
}

// dispatchSync — testable inner. Background goroutine'у нужен свой context
// (request ctx может быть canceled).
func (s *Service) dispatchSync(userID uuid.UUID, event string, payload any) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rows, err := s.Pool.Query(ctx, `
		SELECT id, url, secret, format FROM webhooks
		WHERE user_id = $1 AND active = true AND $2 = ANY(events)`, userID, event)
	if err != nil {
		s.Logger.Error("webhook lookup failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type target struct {
		id     uuid.UUID
		url    string
		secret string
		format string
	}
	targets := []target{}
	for rows.Next() {
		var t target
		if err := rows.Scan(&t.id, &t.url, &t.secret, &t.format); err != nil {
			continue
		}
		targets = append(targets, t)
	}

	// Fan-out parallel delivery: каждый target — отдельная goroutine.
	// errgroup с SetLimit(8) bounds concurrency — типичный user имеет 1-2
	// webhooks, но enterprise team может иметь N штук; не хотим spawn'ить
	// unbounded если N большое + не хотим блокировать одну slow delivery
	// другие. Errors мы игнорируем (logged inside deliver), но errgroup
	// нужен для Wait().
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for _, t := range targets {
		g.Go(func() error {
			s.deliver(gctx, t.id, t.url, t.secret, t.format, event, payload)
			return nil
		})
	}
	_ = g.Wait()
}

// deliver — одна доставка с retry. Backoff 1s/3s/9s на 5xx или network.
// format = "raw" → канонический JSON; "slack" → Slack Block Kit payload.
func (s *Service) deliver(ctx context.Context, id uuid.UUID, url, secret, format, event string, payload any) {
	body, err := formatPayload(Format(format), event, payload)
	if err != nil {
		s.Logger.Error("format payload", zap.String("format", format), zap.Error(err))
		return
	}
	sig := signPayload(secret, body)

	delays := []time.Duration{0, 1 * time.Second, 3 * time.Second, 9 * time.Second}
	var lastStatus int
	for attempt, delay := range delays {
		if delay > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}
		status, err := s.send(ctx, url, sig, event, body)
		lastStatus = status
		if err == nil && status < 500 {
			// 2xx/3xx/4xx — не retry'им, receiver responsibility.
			break
		}
		s.Logger.Debug("webhook retry",
			zap.String("url", url),
			zap.Int("status", status),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		if err != nil {
			lastStatus = -1
		}
	}

	_, _ = s.Pool.Exec(ctx,
		`UPDATE webhooks SET last_delivery_at = now(), last_status = $1 WHERE id = $2`,
		lastStatus, id,
	)
}

// send — единичный POST. Возвращает status и network error если был.
func (s *Service) send(ctx context.Context, url, sig, event string, body []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return -1, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-EoP-Signature", "sha256="+sig)
	req.Header.Set("X-EoP-Event", event)
	req.Header.Set("User-Agent", "eop-webhooks/1.0")

	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()
	return res.StatusCode, nil
}

// signPayload — HMAC-SHA256 над raw JSON body.
func signPayload(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature — для receiver-side тестов и docs/example.
func VerifySignature(secret string, body []byte, header string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	got, err := hex.DecodeString(header[len(prefix):])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}

func generateSecret() (string, error) {
	b := make([]byte, 32) // 256-bit HMAC secret
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "whk_" + hex.EncodeToString(b), nil
}

// Get — single webhook (для UI re-show). Не возвращает secret.
func (s *Service) Get(ctx context.Context, userID, id uuid.UUID) (*Webhook, error) {
	var w Webhook
	err := s.Pool.QueryRow(ctx, `
		SELECT id, url, events, format, active, last_delivery_at, last_status, created_at
		FROM webhooks WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&w.ID, &w.URL, &w.Events, &w.Format, &w.Active, &w.LastDeliveryAt, &w.LastStatus, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}
