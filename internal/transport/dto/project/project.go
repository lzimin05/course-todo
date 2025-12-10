package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProjectDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type PostProjectDTO struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateProjectDTO struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type ProjectMemberDTO struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type AddMemberDTO struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}
