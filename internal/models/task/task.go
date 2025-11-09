package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusWaiting    string = "waiting"
	StatusInProgress string = "in_progress"
	StatusCompleted  string = "completed"
)

type Task struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string 
	Importance  int
	Deadline    time.Time
	CreatedAt   time.Time
	Status      string
}
