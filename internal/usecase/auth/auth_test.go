package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/user"
	"github.com/lzimin05/course-todo/internal/transport/jwt"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func TestAuthUsecase_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockTokenator := mocks.NewMockITokenator(ctrl)
	mockRedisRepo := mocks.NewMockIAuthRedisRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)

	uc := New(mockRepo, mockTokenator, mockRedisRepo, mockProjectRepo)

	tests := []struct {
		name          string
		emailOrLogin  string
		password      string
		setupMocks    func()
		expectedToken string
		expectedError error
	}{
		{
			name:         "successful authentication",
			emailOrLogin: "test@example.com",
			password:     "password123",
			setupMocks: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{
					ID:           uuid.New(),
					Login:        "testuser",
					Username:     "Test User",
					Email:        "test@example.com",
					PasswordHash: hashedPassword,
				}

				mockRepo.EXPECT().
					GetUserByEmailOrLogin(gomock.Any(), "test@example.com").
					Return(user, nil)

				mockTokenator.EXPECT().
					CreateJWT(user.ID.String()).
					Return("test-token", nil)
			},
			expectedToken: "test-token",
			expectedError: nil,
		},
		{
			name:         "user not found",
			emailOrLogin: "nonexistent@example.com",
			password:     "password123",
			setupMocks: func() {
				mockRepo.EXPECT().
					GetUserByEmailOrLogin(gomock.Any(), "nonexistent@example.com").
					Return(nil, errors.New("user not found"))
			},
			expectedToken: "",
			expectedError: errors.New("user not found"),
		},
		{
			name:         "invalid password",
			emailOrLogin: "test@example.com",
			password:     "wrongpassword",
			setupMocks: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{
					ID:           uuid.New(),
					Login:        "testuser",
					Username:     "Test User",
					Email:        "test@example.com",
					PasswordHash: hashedPassword,
				}

				mockRepo.EXPECT().
					GetUserByEmailOrLogin(gomock.Any(), "test@example.com").
					Return(user, nil)
			},
			expectedToken: "",
			expectedError: errors.New("invalid password"),
		},
		{
			name:         "token creation error",
			emailOrLogin: "test@example.com",
			password:     "password123",
			setupMocks: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{
					ID:           uuid.New(),
					Login:        "testuser",
					Username:     "Test User",
					Email:        "test@example.com",
					PasswordHash: hashedPassword,
				}

				mockRepo.EXPECT().
					GetUserByEmailOrLogin(gomock.Any(), "test@example.com").
					Return(user, nil)

				mockTokenator.EXPECT().
					CreateJWT(user.ID.String()).
					Return("", errors.New("token creation failed"))
			},
			expectedToken: "",
			expectedError: errors.New("token creation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			token, err := uc.Authenticate(context.Background(), tt.emailOrLogin, tt.password)

			assert.Equal(t, tt.expectedToken, token)
			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == bcrypt.ErrMismatchedHashAndPassword {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUsecase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockTokenator := mocks.NewMockITokenator(ctrl)
	mockRedisRepo := mocks.NewMockIAuthRedisRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)

	uc := New(mockRepo, mockTokenator, mockRedisRepo, mockProjectRepo)

	tests := []struct {
		name          string
		login         string
		username      string
		email         string
		password      string
		setupMocks    func()
		expectedToken string
		expectedError error
	}{
		{
			name:     "successful registration",
			login:    "testuser",
			username: "Test User",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() {
				userID := uuid.New()
				user := &models.User{
					ID:       userID,
					Login:    "testuser",
					Username: "Test User",
					Email:    "test@example.com",
				}

				mockRepo.EXPECT().
					CreateUser(gomock.Any(), "testuser", "Test User", "test@example.com", gomock.Any()).
					Return(user, nil)

				mockTokenator.EXPECT().
					CreateJWT(userID.String()).
					Return("test-token", nil)

				mockProjectRepo.EXPECT().
					CreateProject(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedToken: "test-token",
			expectedError: nil,
		},
		{
			name:     "duplicate user",
			login:    "existinguser",
			username: "Existing User",
			email:    "existing@example.com",
			password: "password123",
			setupMocks: func() {
				mockRepo.EXPECT().
					CreateUser(gomock.Any(), "existinguser", "Existing User", "existing@example.com", gomock.Any()).
					Return(nil, errs.ErrIsDuplicateKey)
			},
			expectedToken: "",
			expectedError: errs.ErrIsDuplicateKey,
		},
		{
			name:     "database error",
			login:    "testuser",
			username: "Test User",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func() {
				mockRepo.EXPECT().
					CreateUser(gomock.Any(), "testuser", "Test User", "test@example.com", gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedToken: "",
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			token, err := uc.Register(context.Background(), tt.login, tt.username, tt.email, tt.password)

			assert.Equal(t, tt.expectedToken, token)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUsecase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockTokenator := mocks.NewMockITokenator(ctrl)
	mockRedisRepo := mocks.NewMockIAuthRedisRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)

	uc := New(mockRepo, mockTokenator, mockRedisRepo, mockProjectRepo)

	tests := []struct {
		name        string
		token       string
		setupMocks  func()
		expectedErr error
	}{
		{
			name:  "successful logout",
			token: "valid-token",
			setupMocks: func() {
				claims := &jwt.JWTClaims{
					UserID: uuid.New().String(),
				}

				mockTokenator.EXPECT().
					ParseJWT("valid-token").
					Return(claims, nil)

				mockRedisRepo.EXPECT().
					AddToBlacklist(gomock.Any(), claims.UserID, "valid-token").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			setupMocks: func() {
				mockTokenator.EXPECT().
					ParseJWT("invalid-token").
					Return(nil, errors.New("invalid token"))
			},
			expectedErr: errors.New("invalid token"),
		},
		{
			name:  "redis error",
			token: "valid-token",
			setupMocks: func() {
				claims := &jwt.JWTClaims{
					UserID: uuid.New().String(),
				}

				mockTokenator.EXPECT().
					ParseJWT("valid-token").
					Return(claims, nil)

				mockRedisRepo.EXPECT().
					AddToBlacklist(gomock.Any(), claims.UserID, "valid-token").
					Return(errors.New("redis error"))
			},
			expectedErr: errors.New("redis error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := uc.Logout(context.Background(), tt.token)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockTokenator := mocks.NewMockITokenator(ctrl)
	mockRedisRepo := mocks.NewMockIAuthRedisRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)

	uc := New(mockRepo, mockTokenator, mockRedisRepo, mockProjectRepo)

	assert.NotNil(t, uc)
	assert.Equal(t, mockRepo, uc.repo)
	assert.Equal(t, mockTokenator, uc.tokenator)
	assert.Equal(t, mockRedisRepo, uc.redisRepo)
	assert.Equal(t, mockProjectRepo, uc.projectRepo)
}
