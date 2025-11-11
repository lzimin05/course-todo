package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/models/domains"
	"github.com/lzimin05/course-todo/internal/models/errs"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/auth"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUsecase, cfg)

	tests := []struct {
		name           string
		requestBody    dto.LoginRequest
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: dto.LoginRequest{
				EmailOrLogin: "test@example.com",
				Password:     "password123",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Authenticate(gomock.Any(), "test@example.com", "password123").
					Return("test-token", nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Проверяем, что установлена кука
				cookies := w.Result().Cookies()
				require.Len(t, cookies, 1)
				assert.Equal(t, "token", cookies[0].Name)
				assert.Equal(t, "test-token", cookies[0].Value)
				assert.True(t, cookies[0].HttpOnly)
				assert.Equal(t, "/", cookies[0].Path)
			},
		},
		{
			name: "invalid credentials",
			requestBody: dto.LoginRequest{
				EmailOrLogin: "test@example.com",
				Password:     "wrongpassword",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Authenticate(gomock.Any(), "test@example.com", "wrongpassword").
					Return("", fmt.Errorf("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "Incorrect data", response["message"])
			},
		},
		{
			name: "empty email",
			requestBody: dto.LoginRequest{
				EmailOrLogin: "",
				Password:     "password123",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Authenticate(gomock.Any(), "", "password123").
					Return("", fmt.Errorf("empty email"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
		{
			name: "empty password",
			requestBody: dto.LoginRequest{
				EmailOrLogin: "test@example.com",
				Password:     "",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Authenticate(gomock.Any(), "test@example.com", "").
					Return("", fmt.Errorf("empty password"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Добавляем логгер в контекст
			ctx := logctx.WithLogger(req.Context(), logctx.NewLogger())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUsecase, cfg)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	ctx := logctx.WithLogger(req.Context(), logctx.NewLogger())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUsecase, cfg)

	tests := []struct {
		name           string
		requestBody    dto.RegisterRequest
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			requestBody: dto.RegisterRequest{
				Login:    "testuser",
				Username: "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Register(gomock.Any(), "testuser", "Test User", "test@example.com", "password123").
					Return("test-token", nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Проверяем, что установлена кука
				cookies := w.Result().Cookies()
				require.Len(t, cookies, 1)
				assert.Equal(t, "token", cookies[0].Name)
				assert.Equal(t, "test-token", cookies[0].Value)
				assert.True(t, cookies[0].HttpOnly)
			},
		},
		{
			name: "duplicate user",
			requestBody: dto.RegisterRequest{
				Login:    "existinguser",
				Username: "Existing User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Register(gomock.Any(), "existinguser", "Existing User", "existing@example.com", "password123").
					Return("", errs.ErrIsDuplicateKey)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "user with this login or email already exists", response["message"])
			},
		},
		{
			name: "registration error",
			requestBody: dto.RegisterRequest{
				Login:    "testuser",
				Username: "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Register(gomock.Any(), "testuser", "Test User", "test@example.com", "password123").
					Return("", fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "error registation", response["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			ctx := logctx.WithLogger(req.Context(), logctx.NewLogger())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.Register(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUsecase, cfg)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	ctx := logctx.WithLogger(req.Context(), logctx.NewLogger())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockUsecase, cfg)

	tests := []struct {
		name           string
		cookie         *http.Cookie
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "successful logout",
			cookie: &http.Cookie{
				Name:  string(domains.TokenCookieName),
				Value: "test-token",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Logout(gomock.Any(), "test-token").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Проверяем, что кука удалена
				cookies := w.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == string(domains.TokenCookieName) {
						assert.Equal(t, "", cookie.Value)
						break
					}
				}
			},
		},
		{
			name:   "missing token cookie",
			cookie: nil,
			setupMock: func() {
				// Никаких вызовов usecase не ожидается
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "JWT token required", response["message"])
			},
		},
		{
			name: "logout error",
			cookie: &http.Cookie{
				Name:  string(domains.TokenCookieName),
				Value: "invalid-token",
			},
			setupMock: func() {
				mockUsecase.EXPECT().
					Logout(gomock.Any(), "invalid-token").
					Return(fmt.Errorf("token not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "token not found", response["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			ctx := logctx.WithLogger(req.Context(), logctx.NewLogger())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.Logout(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)
	cfg := &config.Config{}

	handler := New(mockUsecase, cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecase, handler.uc)
	assert.Equal(t, cfg, handler.config)
}
