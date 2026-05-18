package webhooksapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/webhooks/domain"
)

type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]domain.Webhook, error)
	Create(ctx context.Context, userID uuid.UUID, url string, events []string, format string) (secret string, hook domain.Webhook, err error)
	Delete(ctx context.Context, userID, id uuid.UUID) (bool, error)
}
