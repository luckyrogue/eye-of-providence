package teamflags_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/teamflags"
)

type memStore struct {
	flags map[string]any
	saveN int64
}

func (m *memStore) Load(ctx context.Context, teamID uuid.UUID) (map[string]any, error) {
	if m.flags == nil {
		return map[string]any{}, nil
	}
	return m.flags, nil
}

func (m *memStore) Save(ctx context.Context, teamID uuid.UUID, flagsJSON []byte) (int64, error) {
	m.saveN = 1
	return 1, nil
}

func TestService_Get(t *testing.T) {
	tid := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	st := &memStore{flags: map[string]any{"x": true}}
	svc := teamflags.New(teamflags.Deps{Store: st, Audit: nil})
	got, err := svc.Get(context.Background(), tid)
	if err != nil {
		t.Fatal(err)
	}
	if got["x"] != true {
		t.Fatalf("%v", got)
	}
}

func TestService_Patch_MissingFlags(t *testing.T) {
	svc := teamflags.New(teamflags.Deps{Store: &memStore{}, Audit: nil})
	_, err := svc.Patch(context.Background(), teamflags.RequestMeta{}, uuid.Nil, "", uuid.Nil, nil)
	if err == nil || !is(err, teamflags.ErrMissingFlags) {
		t.Fatalf("got %v", err)
	}
}

func is(err, target error) bool { return err != nil && err == target }
