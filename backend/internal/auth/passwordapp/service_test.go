package passwordapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/passwordapp"
	"golang.org/x/crypto/bcrypt"
)

type fakeReader struct {
	em, dn, hash string
	uid           uuid.UUID
	err           error
}

func (f fakeReader) LookupByEmail(ctx context.Context, email string) (string, string, string, uuid.UUID, error) {
	if f.err != nil {
		return "", "", "", uuid.Nil, f.err
	}
	return f.em, f.dn, f.hash, f.uid, nil
}

func TestVerifyLogin_OK(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	uid := uuid.New()
	s := passwordapp.New(fakeReader{em: "a@b.c", dn: "Alice", hash: string(hash), uid: uid})
	u, err := s.VerifyLogin(context.Background(), "a@b.c", "secret123")
	if err != nil || u.ID != uid || u.Email != "a@b.c" {
		t.Fatalf("u=%+v err=%v", u, err)
	}
}

func TestVerifyLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.MinCost)
	s := passwordapp.New(fakeReader{hash: string(hash), uid: uuid.New()})
	_, err := s.VerifyLogin(context.Background(), "x@y.z", "wrong")
	if !errors.Is(err, passwordapp.ErrInvalidCredentials) {
		t.Fatalf("got %v", err)
	}
}

func TestVerifyLogin_NoDB(t *testing.T) {
	s := passwordapp.New(nil)
	_, err := s.VerifyLogin(context.Background(), "a@b.c", "x")
	if !errors.Is(err, passwordapp.ErrDBNotConfigured) {
		t.Fatalf("got %v", err)
	}
}
