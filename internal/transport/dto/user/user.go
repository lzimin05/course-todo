package dto

import "github.com/google/uuid"

type UserDTO struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}
