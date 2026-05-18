package membersapp

import "errors"

var (
	ErrOwnerRequired    = errors.New("owner required")
	ErrRoleInsufficient = errors.New("role insufficient")
	ErrInvalidRole      = errors.New("invalid role")
	ErrOwnerLimit       = errors.New("owner limit")
	ErrLastOwner        = errors.New("last owner")
	ErrAdminCantRemove  = errors.New("admin cannot remove owner")
)
