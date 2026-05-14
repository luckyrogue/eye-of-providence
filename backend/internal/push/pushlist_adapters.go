package push

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/push/pushlist"
)

type svcPushListReader struct{ svc *Service }

func (w svcPushListReader) ListByUser(ctx context.Context, userID uuid.UUID) ([]pushlist.SubscriptionRow, error) {
	subs, err := w.svc.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]pushlist.SubscriptionRow, 0, len(subs))
	for _, s := range subs {
		out = append(out, pushlist.SubscriptionRow{
			ID: s.ID, Endpoint: s.Endpoint, UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt, LastUsedAt: s.LastUsedAt,
		})
	}
	return out, nil
}

func newPushListService(svc *Service) *pushlist.Service {
	if svc == nil {
		return pushlist.New(nil)
	}
	return pushlist.New(svcPushListReader{svc: svc})
}
