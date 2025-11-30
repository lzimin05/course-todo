package usecase

import (
	"context"

	"github.com/google/uuid"
	models "github.com/lzimin05/course-todo/internal/models/user"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/user"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/helpers"
)

//go:generate mockgen -source=user.go -destination=../mocks/user_mocks.go -package=mocks UserRepository
type UserRepository interface {
	GetUserByID(context.Context, uuid.UUID) (*models.User, error)
	GetUserByEmail(context.Context, string) (*models.User, error)
	GetUserByLogin(context.Context, string) (*models.User, error)
	UpdateUsername(context.Context, uuid.UUID, string) error
}

type UserUsecase struct {
	repo UserRepository
}

func New(repo UserRepository) *UserUsecase {
	return &UserUsecase{
		repo: repo,
	}
}

func (u *UserUsecase) GetMe(ctx context.Context) (*dto.UserDTO, error) {
	const op = "UserUsecase.GetMe"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("get user from repository")
		return nil, err
	}

	userDTO := &dto.UserDTO{
		ID:       user.ID,
		Login:    user.Login,
		Email:    user.Email,
		Username: user.Username,
	}

	return userDTO, nil
}

func (u *UserUsecase) GetUserByEmail(ctx context.Context, email string) (*dto.UserDTO, error) {
	const op = "UserUsecase.GetUserByEmail"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.WithError(err).Error("get user by email from repository")
		return nil, err
	}

	userDTO := &dto.UserDTO{
		ID:       user.ID,
		Login:    user.Login,
		Email:    user.Email,
		Username: user.Username,
	}

	return userDTO, nil
}

func (u *UserUsecase) GetUserByLogin(ctx context.Context, login string) (*dto.UserDTO, error) {
	const op = "UserUsecase.GetUserByLogin"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	user, err := u.repo.GetUserByLogin(ctx, login)
	if err != nil {
		logger.WithError(err).Error("get user by login from repository")
		return nil, err
	}

	userDTO := &dto.UserDTO{
		ID:       user.ID,
		Login:    user.Login,
		Email:    user.Email,
		Username: user.Username,
	}

	return userDTO, nil
}

func (u *UserUsecase) UpdateUsername(ctx context.Context, username string) error {
	const op = "UserUsecase.UpdateUsername"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return err
	}

	err = u.repo.UpdateUsername(ctx, userID, username)
	if err != nil {
		logger.WithError(err).Error("failed to update username in repository")
		return err
	}

	logger.Info("username updated successfully")
	return nil
}
