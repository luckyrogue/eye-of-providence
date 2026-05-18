package devicesapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/devices/domain"
)

type DeviceRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]domain.Device, error)
	Revoke(ctx context.Context, userID, deviceID uuid.UUID) (bool, error)
}
