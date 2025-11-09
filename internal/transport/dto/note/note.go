package dto

import (
	"time"

	"github.com/google/uuid"
)

type NoteDTO struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateOrUpdateNote struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateNoteDTO struct {
	ID uuid.UUID `json:"id"`
}
