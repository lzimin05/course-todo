package dto

import (
	"time"

	"github.com/google/uuid"
)

type TaskDTO struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	Importance  int       `json:"importance" validate:"required, min=1, max=3"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PostTaskDTO struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	Importance  int       `json:"importance" validate:"required, min=1, max=3"`
	Deadline    time.Time `json:"deadline"`
}
