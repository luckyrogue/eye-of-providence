package authapp

import (
	"context"
	"errors"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	auth PasswordAuthenticator
}

type Deps struct {
	Auth PasswordAuthenticator
}

func New(d Deps) *Service {
	return &Service{auth: d.Auth}
}

func (s *Service) Login(ctx context.Context, email, password string) (LoginUser, error) {
	if s.auth == nil {
		return LoginUser{}, ErrInvalidCredentials
	}
	return s.auth.VerifyLogin(ctx, email, password)
}
