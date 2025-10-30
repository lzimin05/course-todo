package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/user"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lib/pq"
)

const (
	createUserQuery = `
		INSERT INTO todo."user" (id, login, username, email, password_hash) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, login, email, username`

	getUserByEmailQuery = `
		SELECT id, login, username, email, password_hash 
		FROM todo."user" 
		WHERE email = $1`

	getUserByEmailOrLoginQuery = `
		SELECT id, login, username, email, password_hash 
		FROM todo."user" 
		WHERE email = $1 or login = $1`
)

type AuthRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(ctx context.Context, login string, username string, email string, passwordHash []byte) (*models.User, error) {
	const op = "AuthRepository.CreateUser"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("email", email)
	user := &models.User{
		ID:           uuid.New(),
		Login:        login,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	err := r.db.QueryRowContext(ctx, createUserQuery,
		user.ID, user.Login, user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.Login, &user.Email, &user.Username)

	if err != nil {
		// Проверяем, является ли ошибка нарушением уникального ограничения
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			logger.WithError(err).Warn("user with this login or email already exists")
			return nil, errs.ErrIsDuplicateKey
		}
		logger.WithError(err).Error("failed to create user")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "AuthRepository.GetUserByEmail"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("email", email)

	var user models.User
	err := r.db.QueryRowContext(ctx, getUserByEmailQuery, email).
		Scan(&user.ID, &user.Login, &user.Username, &user.Email, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found")
			return nil, nil
		}
		logger.WithError(err).Error("failed to get user by email")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (r *AuthRepository) GetUserByEmailOrLogin(ctx context.Context, email string) (*models.User, error) {
	const op = "AuthRepository.GetUserByEmail"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("email", email)

	var user models.User
	err := r.db.QueryRowContext(ctx, getUserByEmailOrLoginQuery, email).
		Scan(&user.ID, &user.Login, &user.Username, &user.Email, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found")
			return nil, nil
		}
		logger.WithError(err).Error("failed to get user by login")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
