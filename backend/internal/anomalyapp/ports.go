package anomalyapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TrendPoint struct {
	Date     string
	Category string
	MS       uint64
}

type TrendSource interface {
	ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error)
	DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]TrendPoint, error)
}

type WebhookDispatcher interface {
	Dispatch(userID uuid.UUID, event string, payload any)
}

type PushSender interface {
	SendToUser(userID uuid.UUID, payload any)
}
