package validation

import (
	"errors"
	"time"
)

func ValidationTask(title string, importance int, deaadline time.Time) error {
	if title == "" {
		return errors.New("title is required")
	}
	if len(title) < 2 || len(title) > 100 {
		return errors.New("title must be between 2 and 100 characters")
	}

	if importance < 1 || importance > 3 {
		return errors.New("importance must be between 1 and 3")
	}

	if !deaadline.IsZero() && deaadline.Before(time.Now()) {
		return errors.New("deadline must be in the future")
	}

	return nil
}
