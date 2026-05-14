package devices

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/devices/devicelist"
)

type pgDeviceListStore struct{ pool *pgxpool.Pool }

func (p pgDeviceListStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]devicelist.DeviceRow, error) {
	devs, err := List(ctx, p.pool, userID)
	if err != nil {
		return nil, err
	}
	out := make([]devicelist.DeviceRow, 0, len(devs))
	for _, d := range devs {
		out = append(out, devicelist.DeviceRow{
			ID: d.ID, Kind: d.Kind, Name: d.Name, Prefix: d.Prefix,
			CreatedAt: d.CreatedAt, LastUsedAt: d.LastUsedAt,
		})
	}
	return out, nil
}

func newDeviceListService(pool *pgxpool.Pool) *devicelist.Service {
	if pool == nil {
		return devicelist.New(nil)
	}
	return devicelist.New(pgDeviceListStore{pool: pool})
}
