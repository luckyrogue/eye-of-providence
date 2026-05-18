package domain

import "github.com/google/uuid"

type TeamID = uuid.UUID
type UserID = uuid.UUID

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)
