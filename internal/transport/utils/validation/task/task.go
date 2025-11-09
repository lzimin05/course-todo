package validation

import (
	"errors"
	"time"
)

func ValidationTask(title string, importance int, deaadline time.Time) error {
	if title == "" {
		return errors.New("Title is required")
	}

	if importance < 1 || importance > 3 {
		return errors.New("Importance must be between 1 and 3")
	}

	if deaadline.Before(time.Now()) {
		return errors.New("Deadline must be in the future")
	}

	return nil
}
