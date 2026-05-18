package passkeysapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/passkeysapp"
)

type fakeRP struct{}

func (fakeRP) ListPasskeys(context.Context, uuid.UUID) ([]passkeysapp.PasskeyRow, error) {
	return []passkeysapp.PasskeyRow{{ID: uuid.New(), CreatedAt: time.Now()}}, nil
}

func (fakeRP) PasskeyCredentialIDForUser(context.Context, uuid.UUID, uuid.UUID) ([]byte, error) {
	return []byte{1}, nil
}

func (fakeRP) DeletePasskey(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type fakeFactors struct{ n int }

func (f fakeFactors) Count(context.Context, uuid.UUID, *uuid.UUID, []byte) (int, error) {
	return f.n, nil
}

func TestList(t *testing.T) {
	svc := passkeysapp.New(passkeysapp.Deps{RP: fakeRP{}})
	rows, err := svc.List(context.Background(), uuid.New())
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}

func TestDeleteLastFactor(t *testing.T) {
	svc := passkeysapp.New(passkeysapp.Deps{RP: fakeRP{}, Factors: fakeFactors{n: 0}})
	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, passkeysapp.ErrLastAuthFactor) {
		t.Fatalf("err=%v", err)
	}
}
