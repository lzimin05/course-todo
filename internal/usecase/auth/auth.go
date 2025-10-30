package usecase

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/user"
	"github.com/lzimin05/course-todo/internal/transport/jwt"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"golang.org/x/crypto/bcrypt"
)

type ITokenator interface {
	CreateJWT(userID string) (string, error)
	ParseJWT(tokenString string) (*jwt.JWTClaims, error)
}

type AuthRepository interface {
	CreateUser(ctx context.Context, login string, username string, email string, passwordHash []byte) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByEmailOrLogin(ctx context.Context, emailOrLogin string) (*models.User, error)
}

type IAuthRedisRepository interface {
	AddToBlacklist(ctx context.Context, userID, token string) error
}

type AuthUsecase struct {
	repo      AuthRepository
	tokenator ITokenator
	redisRepo IAuthRedisRepository
}

func New(repo AuthRepository, tokenator ITokenator, redisRepo IAuthRedisRepository) *AuthUsecase {
	return &AuthUsecase{
		repo:      repo,
		tokenator: tokenator,
		redisRepo: redisRepo,
	}
}

func (uc *AuthUsecase) Authenticate(ctx context.Context, email, password string) (string, error) {
	const op = "AuthUsecase.Authenticate"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("email", email)

	user, err := uc.repo.GetUserByEmailOrLogin(ctx, email)
	if err != nil {
		logger.WithError(err).Warn("failed to get user by email")
		return "", err
	}
	if user == nil {
		logger.Warn("user not found")
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		logger.Warn("invalid password")
		return "", errors.New("invalid password")
	}

	token, err := uc.tokenator.CreateJWT(user.ID.String())
	if err != nil {
		logger.WithError(err).Error("failed to create JWT")
		return "", err
	}

	return token, nil
}

func (uc *AuthUsecase) Register(ctx context.Context, login, username, email, password string) (string, error) {
	const op = "AuthUsecase.Register"

	logger := logctx.GetLogger(ctx).WithField("op", op).WithFields(map[string]interface{}{
		"email": email,
	})

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.WithError(err).Error("failed to hash password")
		return "", err
	}

	user, err := uc.repo.CreateUser(ctx, login, username, email, hashedPassword)
	if err != nil {
		if err == errs.ErrIsDuplicateKey {
			logger.WithError(err).Error("user with this login or email already exists")
			return "", err
		}
		logger.WithError(err).Error("failed to create user")
		return "", err
	}

	token, err := uc.tokenator.CreateJWT(user.ID.String())
	if err != nil {
		logger.WithError(err).Error("failed to create JWT after registration")
		return "", err
	}

	return token, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, token string) error {
	const op = "AuthUsecase.Logout"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	claims, err := u.tokenator.ParseJWT(token)
	if err != nil {
		logger.WithError(err).Error("failed to parse token")
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidToken)
	}

	if err := u.redisRepo.AddToBlacklist(ctx, claims.UserID, token); err != nil {
		logger.WithError(err).Error("failed to add token to blacklist")
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
