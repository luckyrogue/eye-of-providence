package webhooks

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/webhooks/domain"
	"github.com/eye-of-providence/backend/internal/webhooks/webhooksapp"
)

type repoAdapter struct{ svc *Service }

func (a repoAdapter) List(ctx context.Context, userID uuid.UUID) ([]domain.Webhook, error) {
	hooks, err := a.svc.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Webhook, len(hooks))
	for i := range hooks {
		out[i] = domain.Webhook(hooks[i])
	}
	return out, nil
}

func (a repoAdapter) Create(ctx context.Context, userID uuid.UUID, url string, events []string, format string) (string, domain.Webhook, error) {
	secret, hook, err := a.svc.Create(ctx, userID, url, events, format)
	return secret, domain.Webhook(hook), err
}

func (a repoAdapter) Delete(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	return a.svc.Delete(ctx, userID, id)
}

func newWebhooksApp(svc *Service) *webhooksapp.Service {
	if svc == nil {
		return webhooksapp.New(webhooksapp.Deps{})
	}
	return webhooksapp.New(webhooksapp.Deps{Repo: repoAdapter{svc: svc}})
}
