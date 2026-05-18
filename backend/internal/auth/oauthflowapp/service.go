package oauthflowapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/eye-of-providence/backend/internal/auth/oauthapp"
)

type Service struct {
	linker  UserLinker
	session SessionIssuer
}

type Deps struct {
	Linker  UserLinker
	Session SessionIssuer
}

func New(d Deps) *Service {
	return &Service{linker: d.Linker, session: d.Session}
}

func (s *Service) RandomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Service) StateCookieValue(nonce, returnTo string) string {
	if returnTo == "" {
		return nonce
	}
	return nonce + "|" + returnTo
}

func (s *Service) ParseStoredState(stored string) (nonce, returnTo string) {
	nonce = stored
	if i := strings.Index(stored, "|"); i > 0 {
		nonce = stored[:i]
		returnTo = stored[i+1:]
	}
	return nonce, returnTo
}

type CallbackInput struct {
	GotState          string
	StoredStateCookie string
	Code              string
	OAuthError        string
}

type CallbackResult struct {
	Token    string
	ReturnTo string
}

func (s *Service) CompleteCallback(ctx context.Context, provider string, prov Provider, in CallbackInput) (CallbackResult, error) {
	nonce, returnTo := s.ParseStoredState(in.StoredStateCookie)
	if in.GotState == "" || in.GotState != nonce {
		return CallbackResult{}, ErrStateMismatch
	}
	if in.OAuthError == "access_denied" {
		return CallbackResult{}, ErrUserDenied
	}
	if in.Code == "" {
		return CallbackResult{}, ErrMissingCode
	}
	ext, err := prov.Exchange(ctx, in.Code)
	if err != nil {
		return CallbackResult{}, errors.Join(ErrOAuthExchangeFailed, err)
	}
	if ext.Email == "" {
		return CallbackResult{}, ErrEmailNotVerified
	}
	userID, err := s.linker.UpsertOAuthUser(ctx, provider, ext)
	if err != nil {
		if errors.Is(err, oauthapp.ErrIdentityLinkRequired) {
			return CallbackResult{}, &IdentityLinkConflict{Email: ext.Email, Provider: provider}
		}
		return CallbackResult{}, err
	}
	tok, err := s.session.IssueHandoff(ctx, userID, ext.Email, provider)
	if err != nil {
		return CallbackResult{}, err
	}
	return CallbackResult{Token: tok, ReturnTo: returnTo}, nil
}
