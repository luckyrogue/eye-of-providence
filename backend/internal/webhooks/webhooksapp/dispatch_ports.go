package webhooksapp

import (
	"context"

	"github.com/google/uuid"
)

type DeliveryTarget struct {
	ID     uuid.UUID
	URL    string
	Secret string
	Format string
}

type ActiveWebhookLister interface {
	ListActive(ctx context.Context, userID uuid.UUID, event string) ([]DeliveryTarget, error)
}

type DeliveryRecorder interface {
	Record(ctx context.Context, webhookID uuid.UUID, status int) error
}

type PayloadFormatter interface {
	Format(format, event string, payload any) ([]byte, error)
}

type HTTPPoster interface {
	Post(ctx context.Context, url string, headers map[string]string, body []byte) (status int, err error)
}

type Signer interface {
	Sign(secret string, body []byte) string
}
