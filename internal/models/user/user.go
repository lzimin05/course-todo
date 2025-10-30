package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
}
