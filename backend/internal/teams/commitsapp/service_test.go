package commitsapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/commitsapp"
	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type fakeIngest struct{}

func (fakeIngest) Ingest(context.Context, uuid.UUID, domain.CommitInput) (bool, uuid.UUID, uuid.UUID, error) {
	return true, uuid.New(), uuid.New(), nil
}

func TestIngest(t *testing.T) {
	ok, _, _, err := commitsapp.New(commitsapp.Deps{Ingest: fakeIngest{}}).Ingest(context.Background(), uuid.New(), domain.CommitInput{AuthoredAt: time.Now()})
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
