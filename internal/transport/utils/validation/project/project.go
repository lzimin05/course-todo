package validation

import "errors"

func ValidationProject(name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) < 2 || len(name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	return nil
}
