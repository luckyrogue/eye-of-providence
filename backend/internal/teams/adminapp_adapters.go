package teams

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/store"
	"github.com/eye-of-providence/backend/internal/teams/adminapp"
)

func (s *Service) adminApp() *adminapp.Service {
	var deleter adminapp.UserDeleter
	if d, ok := s.EventStore.(store.UserDeleter); ok {
		deleter = userDeleterAdapter{d: d, log: s.Logger}
	}
	return adminapp.New(adminapp.Deps{
		Store:       adminapp.NewPGStore(s.Pool),
		Audit:       s.Audit,
		UserDeleter: deleter,
		TokenBumper: tokenBumperAdapter{pool: s.Pool},
		Users:       userFinderAdapter{pool: s.Pool},
	})
}

type userDeleterAdapter struct {
	d   store.UserDeleter
	log *zap.Logger
}

func (a userDeleterAdapter) DeleteUserData(ctx context.Context, userID string) error {
	if err := a.d.DeleteUserData(ctx, userID); err != nil && a.log != nil {
		a.log.Warn("clickhouse delete failed (continuing)",
			zap.String("user_id", userID), zap.Error(err))
	}
	return nil
}

type tokenBumperAdapter struct {
	pool *pgxpool.Pool
}

func (a tokenBumperAdapter) BumpTokenVersion(ctx context.Context, userID uuid.UUID) error {
	return auth.BumpTokenVersion(ctx, a.pool, userID)
}

type userFinderAdapter struct {
	pool *pgxpool.Pool
}

func (a userFinderAdapter) FindByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	u, err := auth.FindUserByEmail(ctx, a.pool, email)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return uuid.Nil, adminapp.ErrUserNotFound
		}
		return uuid.Nil, err
	}
	return u.ID, nil
}
