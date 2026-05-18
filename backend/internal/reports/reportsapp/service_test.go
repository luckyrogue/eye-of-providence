package reportsapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/reports/domain"
	"github.com/eye-of-providence/backend/internal/reports/reportsapp"
)

type memRepo struct {
	items []domain.Report
}

func (m *memRepo) Save(r domain.Report) { m.items = append(m.items, r) }

func (m *memRepo) ListForUser(userID string, limit int) []domain.Report {
	var out []domain.Report
	for i := len(m.items) - 1; i >= 0 && len(out) < limit; i-- {
		if m.items[i].UserID == userID {
			out = append(out, m.items[i])
		}
	}
	return out
}

func (m *memRepo) Get(id, userID string) (domain.Report, bool) {
	for _, r := range m.items {
		if r.ID == id && r.UserID == userID {
			return r, true
		}
	}
	return domain.Report{}, false
}

type fakeGen struct{}

func (fakeGen) Generate(context.Context, *domain.NumericContext) (string, error) { return "# ok", nil }
func (fakeGen) Model() string                                                   { return "test" }

type fakeBuilder struct{}

func (fakeBuilder) Build(context.Context, string, string, time.Time, time.Time) (*domain.NumericContext, error) {
	return &domain.NumericContext{}, nil
}

func TestGenerate(t *testing.T) {
	repo := &memRepo{}
	svc := reportsapp.New(reportsapp.Deps{Repo: repo, Builder: fakeBuilder{}, Generator: fakeGen{}})
	r, err := svc.Generate(context.Background(), "u1", "weekly", time.Now().UTC())
	if err != nil || r.UserID != "u1" || r.BodyMD == "" {
		t.Fatalf("r=%+v err=%v", r, err)
	}
}
