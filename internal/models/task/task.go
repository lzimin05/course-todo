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
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Importance  int       `json:"importance"`
	Deadline    time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}
