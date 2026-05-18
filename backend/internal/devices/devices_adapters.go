package devices

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/devices/devicesapp"
	"github.com/eye-of-providence/backend/internal/devices/domain"
	"github.com/eye-of-providence/backend/internal/devices/pairingapp"
)

type deviceRepoAdapter struct{ pool *pgxpool.Pool }

func (a deviceRepoAdapter) List(ctx context.Context, userID uuid.UUID) ([]domain.Device, error) {
	devs, err := List(ctx, a.pool, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Device, len(devs))
	for i := range devs {
		out[i] = domain.Device(devs[i])
	}
	return out, nil
}

func (a deviceRepoAdapter) Revoke(ctx context.Context, userID, deviceID uuid.UUID) (bool, error) {
	return Revoke(ctx, a.pool, userID, deviceID)
}

type pairingRepoAdapter struct{ pool *pgxpool.Pool }

func (a pairingRepoAdapter) Begin(ctx context.Context, kind, nameHint string) (domain.PairBeginResult, error) {
	r, err := PairBegin(ctx, a.pool, kind, nameHint)
	return domain.PairBeginResult(r), err
}

func (a pairingRepoAdapter) Poll(ctx context.Context, pairID uuid.UUID, secret string) (domain.PollResult, error) {
	r, err := Poll(ctx, a.pool, pairID, secret)
	return domain.PollResult(r), err
}

func (a pairingRepoAdapter) Claim(ctx context.Context, userID uuid.UUID, code, name string) (domain.Device, error) {
	d, err := Claim(ctx, a.pool, userID, code, name)
	return domain.Device(d), err
}

func newDevicesApp(pool *pgxpool.Pool) *devicesapp.Service {
	if pool == nil {
		return devicesapp.New(devicesapp.Deps{})
	}
	return devicesapp.New(devicesapp.Deps{Repo: deviceRepoAdapter{pool: pool}})
}

func newPairingApp(pool *pgxpool.Pool) *pairingapp.Service {
	if pool == nil {
		return pairingapp.New(pairingapp.Deps{})
	}
	return pairingapp.New(pairingapp.Deps{Repo: pairingRepoAdapter{pool: pool}})
}
