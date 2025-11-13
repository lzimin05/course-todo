package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/models/domains"
	"github.com/lzimin05/course-todo/internal/models/errs"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/user"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func TestUserTransport_GetMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUsecase := mocks.NewMockIUserUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUserUsecase, cfg)

	tests := []struct {
		name       string
		mockFunc   func()
		statusCode int
	}{
		{
			name: "Success",
			mockFunc: func() {
				userDTO := &dto.UserDTO{
					ID:       uuid.New(),
					Login:    "testuser",
					Username: "Test User",
					Email:    "test@example.com",
				}
				mockUserUsecase.EXPECT().GetMe(gomock.Any()).Return(userDTO, nil)
			},
			statusCode: http.StatusOK,
		},
		{
			name: "User not found",
			mockFunc: func() {
				mockUserUsecase.EXPECT().GetMe(gomock.Any()).Return(nil, errs.NewNotFoundError("user not found"))
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetMe(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestUserTransport_GetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUsecase := mocks.NewMockIUserUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUserUsecase, cfg)

	tests := []struct {
		name       string
		email      string
		mockFunc   func()
		statusCode int
	}{
		{
			name:  "Success",
			email: "test@example.com",
			mockFunc: func() {
				userDTO := &dto.UserDTO{
					ID:       uuid.New(),
					Login:    "testuser",
					Username: "Test User",
					Email:    "test@example.com",
				}
				mockUserUsecase.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(userDTO, nil)
			},
			statusCode: http.StatusOK,
		},
		{
			name:  "Empty email parameter",
			email: "",
			mockFunc: func() {
				// No expectation as handler returns early
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:  "User not found",
			email: "notfound@example.com",
			mockFunc: func() {
				mockUserUsecase.EXPECT().GetUserByEmail(gomock.Any(), "notfound@example.com").Return(nil, errs.NewNotFoundError("user not found"))
			},
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			url := "/users/by-email"
			if tt.email != "" {
				url += "?email=" + tt.email
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetUserByEmail(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestUserTransport_GetUserByLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUsecase := mocks.NewMockIUserUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUserUsecase, cfg)

	tests := []struct {
		name       string
		login      string
		mockFunc   func()
		statusCode int
	}{
		{
			name:  "Success",
			login: "testuser",
			mockFunc: func() {
				userDTO := &dto.UserDTO{
					ID:       uuid.New(),
					Login:    "testuser",
					Username: "Test User",
					Email:    "test@example.com",
				}
				mockUserUsecase.EXPECT().GetUserByLogin(gomock.Any(), "testuser").Return(userDTO, nil)
			},
			statusCode: http.StatusOK,
		},
		{
			name:  "Empty login parameter",
			login: "",
			mockFunc: func() {
				// No expectation as handler returns early
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:  "User not found",
			login: "notfound",
			mockFunc: func() {
				mockUserUsecase.EXPECT().GetUserByLogin(gomock.Any(), "notfound").Return(nil, errs.NewNotFoundError("user not found"))
			},
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			url := "/users/by-login"
			if tt.login != "" {
				url += "?login=" + tt.login
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetUserByLogin(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}
