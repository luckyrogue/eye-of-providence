package teams

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/teams/registrationapp"
)

type pgUserCounter struct{ s *Service }

func (p pgUserCounter) UserCount(ctx context.Context) (int, error) {
	var n int
	err := p.s.Pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&n)
	return n, err
}

type pgFirstUserPromoter struct{ s *Service }

func (p pgFirstUserPromoter) PromoteSuperAdmin(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	_, err = p.s.Pool.Exec(ctx, "UPDATE users SET global_role='super_admin' WHERE id=$1", uid)
	if err == nil && p.s.Logger != nil {
		p.s.Logger.Info("first user promoted to super_admin", zap.String("user_id", userID))
	}
	return err
}

func (s *Service) registrationApp() *registrationapp.Service {
	return registrationapp.New(registrationapp.Deps{
		Users: pgUserCounter{s: s},
		Promo: pgFirstUserPromoter{s: s},
	})
}
