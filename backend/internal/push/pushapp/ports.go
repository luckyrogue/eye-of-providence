package pushapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/push/domain"
)

type Repository interface {
	List(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error)
	Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth, userAgent string) error
	Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) (bool, error)
}
