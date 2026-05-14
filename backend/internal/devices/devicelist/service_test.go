package devicelist_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/devices/devicelist"
)

type fakeStore struct {
	rows []devicelist.DeviceRow
	err  error
}

func (f fakeStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]devicelist.DeviceRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]devicelist.DeviceRow(nil), f.rows...), nil
}

func TestListMyDevices_NilStore(t *testing.T) {
	s := devicelist.New(nil)
	got, err := s.ListMyDevices(context.Background(), uuid.New())
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v err=%v", got, err)
	}
}

func TestListMyDevices_Delegates(t *testing.T) {
	uid := uuid.New()
	row := devicelist.DeviceRow{ID: uid, Kind: "ext", Name: "x", Prefix: "eop_"}
	s := devicelist.New(fakeStore{rows: []devicelist.DeviceRow{row}})
	got, err := s.ListMyDevices(context.Background(), uid)
	if err != nil || len(got) != 1 || got[0].Kind != "ext" {
		t.Fatalf("%+v err=%v", got, err)
	}
}
