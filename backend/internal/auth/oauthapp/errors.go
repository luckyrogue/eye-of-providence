package oauthapp

import "errors"

var ErrIdentityLinkRequired = errors.New("identity link requires re-auth")
