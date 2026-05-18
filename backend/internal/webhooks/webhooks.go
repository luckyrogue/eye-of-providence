package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/plans"
)

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

type Webhook struct {
	ID             uuid.UUID  `json:"id"`
	URL            string     `json:"url"`
	Events         []string   `json:"events"`
	Format         string     `json:"format"`
	Active         bool       `json:"active"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	LastStatus     *int       `json:"last_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Service struct {
	Pool       *pgxpool.Pool
	Logger     *zap.Logger
	HTTPClient *http.Client
	Plans      plans.Service

	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc

	inflight sync.WaitGroup

	dispatchSem chan struct{}
}

const maxConcurrentDispatch = 64

func New(pool *pgxpool.Pool, logger *zap.Logger) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		Pool:           pool,
		Logger:         logger,
		HTTPClient:     &http.Client{Timeout: 5 * time.Second},
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
		dispatchSem:    make(chan struct{}, maxConcurrentDispatch),
	}
}

func (s *Service) Shutdown(ctx context.Context) error {
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

func (s *Service) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM webhooks WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Service) Dispatch(userID uuid.UUID, event string, payload any) {
	if !validEvents[event] {
		s.Logger.Warn("unknown event for dispatch", zap.String("event", event))
		return
	}

	select {
	case s.dispatchSem <- struct{}{}:
	case <-s.shutdownCtx.Done():

		return
	}
	s.inflight.Add(1)
	go func() {
		defer func() {
			<-s.dispatchSem
			s.inflight.Done()
		}()
		s.dispatchSync(userID, event, payload)
	}()
}

func (s *Service) dispatchSync(userID uuid.UUID, event string, payload any) {
	parent := s.shutdownCtx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()
	s.newDispatcher().DispatchSync(ctx, userID, event, payload)
}

func signPayload(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

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
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "whk_" + hex.EncodeToString(b), nil
}

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
