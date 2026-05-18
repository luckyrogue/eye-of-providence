package oauthflowapp

import "errors"

var (
	ErrStateMismatch         = errors.New("state mismatch")
	ErrUserDenied            = errors.New("user denied")
	ErrMissingCode           = errors.New("missing code")
	ErrEmailNotVerified      = errors.New("email not verified")
	ErrOAuthExchangeFailed   = errors.New("oauth exchange failed")
	ErrIdentityLinkRequired  = errors.New("identity link required")
)

type IdentityLinkConflict struct {
	Email    string
	Provider string
}

func (e *IdentityLinkConflict) Error() string { return ErrIdentityLinkRequired.Error() }

func (e *IdentityLinkConflict) Is(target error) bool {
	return target == ErrIdentityLinkRequired
}
