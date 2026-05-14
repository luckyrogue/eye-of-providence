package webhooks

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/webhooks/webhooklist"
)

type svcWebhookListReader struct{ svc *Service }

func (w svcWebhookListReader) ListByUser(ctx context.Context, userID uuid.UUID) ([]webhooklist.Row, error) {
	hooks, err := w.svc.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]webhooklist.Row, 0, len(hooks))
	for _, h := range hooks {
		out = append(out, webhooklist.Row{
			ID: h.ID, URL: h.URL, Events: h.Events, Format: h.Format, Active: h.Active,
			LastDeliveryAt: h.LastDeliveryAt, LastStatus: h.LastStatus, CreatedAt: h.CreatedAt,
		})
	}
	return out, nil
}

func newWebhookListService(svc *Service) *webhooklist.Service {
	if svc == nil {
		return webhooklist.New(nil)
	}
	return webhooklist.New(svcWebhookListReader{svc: svc})
}
