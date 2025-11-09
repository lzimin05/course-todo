package models

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         string
	Description  string
	CreatedAt    time.Time
}