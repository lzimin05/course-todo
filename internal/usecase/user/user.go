package usecase

import (
	"context"

	"github.com/google/uuid"
	models "github.com/lzimin05/course-todo/internal/models/user"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/user"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/helpers"
)

type IUserRepository interface {
	GetUserByID(context.Context, uuid.UUID) (*models.User, error)
}

type UserUsecase struct {
	repo IUserRepository
}

func New(repo IUserRepository) *UserUsecase{
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
		Email: user.Email,
		Username: user.Username,
		Login: user.Login,
	}

	return userDTO, nil
}