package validation

import (
	"errors"
	"regexp"
	"strings"

	dto "github.com/lzimin05/course-todo/internal/transport/dto/auth"
)

func ValidateRegisterRequest(req dto.RegisterRequest) error {
	// Валидация логина
	if strings.TrimSpace(req.Login) == "" {
		return errors.New("login is required")
	}
	if len(req.Login) < 3 || len(req.Login) > 50 {
		return errors.New("login must be between 3 and 50 characters")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(req.Login) {
		return errors.New("login can only contain letters, numbers and underscores")
	}

	// Валидация имени пользователя
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 2 || len(req.Username) > 100 {
		return errors.New("username must be between 2 and 100 characters")
	}

	// Валидация email
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email is required")
	}
	if len(req.Email) > 255 {
		return errors.New("email is too long")
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(strings.ToLower(req.Email)) {
		return errors.New("invalid email format")
	}

	// Валидация пароля
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(req.Password) > 72 {
		return errors.New("password is too long")
	}

	return nil
}
