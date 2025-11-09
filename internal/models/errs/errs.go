package models

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvaliidRequest    = errors.New("invalid request")
	ErrNotFound           = errors.New("not found")
	ErrInvalidID          = errors.New("invalid id format")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrIsDuplicateKey     = errors.New("user with this login or email already exists")
	ErrEmptyNoteName      = errors.New("note name cannot be empty")
)

func NewNotFoundError(msg string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, msg)
}
