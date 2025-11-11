package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/user"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/user"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func TestUserUsecase_GetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	uc := New(mockUserRepo)

	tests := []struct {
		name     string
		email    string
		mockFunc func()
		expected *dto.UserDTO
		wantErr  bool
		errMsg   string
	}{
		{
			name:  "Success",
			email: "test@example.com",
			mockFunc: func() {
				user := &models.User{
					ID:       uuid.New(),
					Login:    "testuser",
					Username: "Test User",
					Email:    "test@example.com",
				}
				mockUserRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(user, nil)
			},
			expected: &dto.UserDTO{
				Login:    "testuser",
				Username: "Test User",
				Email:    "test@example.com",
			},
			wantErr: false,
		},
		{
			name:  "User not found",
			email: "notfound@example.com",
			mockFunc: func() {
				mockUserRepo.EXPECT().GetUserByEmail(gomock.Any(), "notfound@example.com").Return(nil, errs.NewNotFoundError("user not found"))
			},
			expected: nil,
			wantErr:  true,
			errMsg:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			result, err := uc.GetUserByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Login, result.Login)
				assert.Equal(t, tt.expected.Username, result.Username)
				assert.Equal(t, tt.expected.Email, result.Email)
			}
		})
	}
}

func TestUserUsecase_GetUserByLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	uc := New(mockUserRepo)

	tests := []struct {
		name     string
		login    string
		mockFunc func()
		expected *dto.UserDTO
		wantErr  bool
		errMsg   string
	}{
		{
			name:  "Success",
			login: "testuser",
			mockFunc: func() {
				user := &models.User{
					ID:       uuid.New(),
					Login:    "testuser",
					Username: "Test User",
					Email:    "test@example.com",
				}
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "testuser").Return(user, nil)
			},
			expected: &dto.UserDTO{
				Login:    "testuser",
				Username: "Test User",
				Email:    "test@example.com",
			},
			wantErr: false,
		},
		{
			name:  "User not found",
			login: "notfound",
			mockFunc: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "notfound").Return(nil, errs.NewNotFoundError("user not found"))
			},
			expected: nil,
			wantErr:  true,
			errMsg:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			result, err := uc.GetUserByLogin(context.Background(), tt.login)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Login, result.Login)
				assert.Equal(t, tt.expected.Username, result.Username)
				assert.Equal(t, tt.expected.Email, result.Email)
			}
		})
	}
}
