package webauthnapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/webauthnapp"
)

type stubRP struct {
	sid string
	err error
}

func (s stubRP) BeginRegistration(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error) {
	if s.err != nil {
		return nil, "", s.err
	}
	return nil, s.sid, nil
}

func (stubRP) FinishRegistration(ctx context.Context, userID uuid.UUID, sid string, body []byte, nickname string) error {
	return nil
}

func (stubRP) BeginLogin(ctx context.Context, email *string) (*protocol.CredentialAssertion, string, error) {
	return nil, "", nil
}

func (stubRP) FinishLogin(ctx context.Context, sid string, body []byte) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func TestRegisterBegin_Forward(t *testing.T) {
	svc := webauthnapp.New(stubRP{sid: "sess-1"})
	_, sid, err := svc.RegisterBegin(context.Background(), uuid.New())
	if err != nil || sid != "sess-1" {
		t.Fatalf("sid=%q err=%v", sid, err)
	}
}

func TestRegisterBegin_Error(t *testing.T) {
	svc := webauthnapp.New(stubRP{err: errors.New("boom")})
	_, _, err := svc.RegisterBegin(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected err")
	}
}
