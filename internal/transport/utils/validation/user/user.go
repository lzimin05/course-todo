package validation

import (
	"errors"
	"strings"

	dto "github.com/lzimin05/course-todo/internal/transport/dto/user"
)

func ValidateUpdateUsernameRequest(req dto.UpdateUsernameRequest) error {
	// Валидация имени пользователя
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 2 || len(req.Username) > 100 {
		return errors.New("username must be between 2 and 100 characters")
	}

	return nil
}