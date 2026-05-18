package teams

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/teams/authapp"
)

type passwordAuthAdapter struct {
	pool *pgxpool.Pool
}

func (a passwordAuthAdapter) VerifyLogin(ctx context.Context, email, password string) (authapp.LoginUser, error) {
	u, err := auth.NewPasswordLoginService(a.pool).VerifyLogin(ctx, email, password)
	if err != nil {
		return authapp.LoginUser{}, err
	}
	return authapp.LoginUser{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName}, nil
}

func (s *Service) authApp() *authapp.Service {
	return authapp.New(authapp.Deps{Auth: passwordAuthAdapter{pool: s.Pool}})
}
