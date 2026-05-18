package invitesapp

import "errors"

var (
	ErrInviteInvalid      = errors.New("invite invalid or expired")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrRoleInsufficient   = errors.New("role insufficient")
	ErrPlanLimitExceeded  = errors.New("plan limit exceeded")
)
