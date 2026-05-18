package push

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/push/domain"
	"github.com/eye-of-providence/backend/internal/push/pushapp"
)

type repoAdapter struct{ svc *Service }

func (a repoAdapter) List(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	subs, err := a.svc.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Subscription, len(subs))
	for i := range subs {
		out[i] = domain.Subscription(subs[i])
	}
	return out, nil
}

func (a repoAdapter) Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth, userAgent string) error {
	return a.svc.Subscribe(ctx, userID, endpoint, p256dh, auth, userAgent)
}

func (a repoAdapter) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) (bool, error) {
	return a.svc.Unsubscribe(ctx, userID, endpoint)
}

func newPushApp(svc *Service) *pushapp.Service {
	if svc == nil {
		return pushapp.New(pushapp.Deps{})
	}
	return pushapp.New(pushapp.Deps{Repo: repoAdapter{svc: svc}})
}
