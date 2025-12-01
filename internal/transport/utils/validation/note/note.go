package validation

import "errors"

func ValidationNote(name string, description string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) < 2 || len(name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	if len(description) > 5000 {
		return errors.New("description must be at most 5000 characters")
	}
	return nil
}
