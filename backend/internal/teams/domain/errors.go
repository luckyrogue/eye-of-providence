package domain

import "errors"

var (
	ErrInviteInvalid    = errors.New("invite invalid or expired")
	ErrNotMember        = errors.New("not a team member")
	ErrProjectOrphaned  = errors.New("project orphaned")
	ErrProjectNotFound  = errors.New("project not found")
)
