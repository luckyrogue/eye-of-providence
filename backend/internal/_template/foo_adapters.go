package templatebc

import (
	"context"

	"github.com/eye-of-providence/backend/internal/_template/domain"
	"github.com/eye-of-providence/backend/internal/_template/fooapp"
)

// pgRepo implements domain.Repository — wire real pgx queries here.
type pgRepo struct{}

func (pgRepo) FindByID(context.Context, string) (*domain.Entity, error) {
	return nil, domain.ErrNotFound
}

func NewService() *fooapp.Service {
	return fooapp.New(fooapp.Deps{Repo: pgRepo{}})
}
