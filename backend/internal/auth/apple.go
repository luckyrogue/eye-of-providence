package auth

import (
	"context"
	"errors"
	"net/url"
)

type AppleOAuth struct {
	TeamID      string
	KeyID       string
	ClientID    string
	PrivateKey  string
	CallbackURL string
}

var _ OAuthProvider = (*AppleOAuth)(nil)

func NewAppleOAuth(teamID, keyID, clientID, privateKey, callbackURL string) *AppleOAuth {
	return &AppleOAuth{
		TeamID:      teamID,
		KeyID:       keyID,
		ClientID:    clientID,
		PrivateKey:  privateKey,
		CallbackURL: callbackURL,
	}
}

func (a *AppleOAuth) Name() string { return "apple" }

func (a *AppleOAuth) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", a.ClientID)
	q.Set("redirect_uri", a.CallbackURL)
	q.Set("response_type", "code")
	q.Set("scope", "name email")
	q.Set("response_mode", "form_post")
	q.Set("state", state)
	return "https://appleid.apple.com/auth/authorize?" + q.Encode()
}

func (a *AppleOAuth) Exchange(_ context.Context, _ string) (*ExternalUser, error) {

	return nil, errors.New("apple sign-in: phase 1 stub, exchange not implemented yet")
}
