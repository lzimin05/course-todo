package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID
	Name        string
	Description string
	OwnerID     uuid.UUID
	CreatedAt   time.Time
}

const (
	RoleOwner  string = "owner"
	RoleMember string = "member"
)

type ProjectMember struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Username  string
	Email     string
	Role      string
	JoinedAt  time.Time
}
