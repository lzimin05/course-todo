package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID
	Login        string 
	Username     string 
	Email        string 
	PasswordHash []byte 
}
