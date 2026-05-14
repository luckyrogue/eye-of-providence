package passwordapp

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginUser struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
}

type Reader interface {
	LookupByEmail(ctx context.Context, email string) (emailOut, displayName, passwordHash string, userID uuid.UUID, err error)
}

type Service struct {
	r Reader
}

func New(r Reader) *Service {
	return &Service{r: r}
}

func (s *Service) VerifyLogin(ctx context.Context, email, password string) (LoginUser, error) {
	if s.r == nil {
		return LoginUser{}, ErrDBNotConfigured
	}
	em, dn, hash, uid, err := s.r.LookupByEmail(ctx, email)
	if err != nil {
		return LoginUser{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return LoginUser{}, ErrInvalidCredentials
	}
	return LoginUser{ID: uid, Email: em, DisplayName: dn}, nil
}
