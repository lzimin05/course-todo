package dto

import "github.com/google/uuid"

type NoteDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type CreateOrUpdateNote struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}