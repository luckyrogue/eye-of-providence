package webhooks

import (
	"bytes"
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/webhooks/webhooksapp"
)

type pgWebhookLister struct {
	pool *pgxpool.Pool
}

func (p pgWebhookLister) ListActive(ctx context.Context, userID uuid.UUID, event string) ([]webhooksapp.DeliveryTarget, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, url, secret, format FROM webhooks
		WHERE user_id = $1 AND active = true AND $2 = ANY(events)`, userID, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []webhooksapp.DeliveryTarget{}
	for rows.Next() {
		var t webhooksapp.DeliveryTarget
		if err := rows.Scan(&t.ID, &t.URL, &t.Secret, &t.Format); err != nil {
			continue
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type pgDeliveryRecorder struct {
	pool *pgxpool.Pool
}

func (p pgDeliveryRecorder) Record(ctx context.Context, webhookID uuid.UUID, status int) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE webhooks SET last_delivery_at = now(), last_status = $1 WHERE id = $2`,
		status, webhookID)
	return err
}

type whFormatAdapter struct{}

func (whFormatAdapter) Format(format, event string, payload any) ([]byte, error) {
	return formatPayload(Format(format), event, payload)
}

type whSignAdapter struct{}

func (whSignAdapter) Sign(secret string, body []byte) string {
	return signPayload(secret, body)
}

type whHTTPPoster struct {
	client *http.Client
}

func (h whHTTPPoster) Post(ctx context.Context, url string, headers map[string]string, body []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return -1, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := h.client.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()
	return res.StatusCode, nil
}

func (s *Service) newDispatcher() *webhooksapp.Dispatcher {
	return &webhooksapp.Dispatcher{
		Lister: pgWebhookLister{pool: s.Pool},
		Format: whFormatAdapter{},
		Sign:   whSignAdapter{},
		HTTP:   whHTTPPoster{client: s.HTTPClient},
		Record: pgDeliveryRecorder{pool: s.Pool},
		Logger: s.Logger,
		ValidEvent: func(event string) bool {
			return validEvents[event]
		},
	}
}
