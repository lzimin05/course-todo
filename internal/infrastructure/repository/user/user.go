package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/user"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

const (
	queryGetUserByID = `
	SELECT 
		u.id, 
		u.login,
		u.username, 
		u.email, 
		u.password_hash
	FROM todo.user u
	WHERE u.id = $1;`
)

type UserRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "UserRepository.GetUserByID"
	logger := logctx.GetLogger(ctx).WithField("op", op)
	
	var user models.User

	err := r.db.QueryRowContext(ctx, queryGetUserByID, id).Scan(
		&user.ID,
		&user.Login,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
	)
	if err != nil {
		logger.WithError(err).Warn("err get user by id")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, err
	}

	return &user, nil
}
