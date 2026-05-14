package oauthapp

import "errors"

// ErrIdentityLinkRequired — email collision: требуется повторная аутентификация (Phase 2+).
var ErrIdentityLinkRequired = errors.New("identity link requires re-auth")
