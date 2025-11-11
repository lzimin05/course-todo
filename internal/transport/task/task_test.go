package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/models/domains"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/task"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func TestTaskTransport_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskUsecase := mocks.NewMockTaskUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockTaskUsecase, cfg)

	tests := []struct {
		name       string
		body       interface{}
		mockFunc   func()
		statusCode int
		wantErr    bool
	}{
		{
			name: "Success",
			body: dto.PostTaskDTO{
				Title:       "Test Task",
				Description: "Test Description",
				ProjectID:   uuid.New(),
				Importance:  1,
				Deadline:    time.Now().Add(24 * time.Hour),
			},
			mockFunc: func() {
				mockTaskUsecase.EXPECT().CreateTask(gomock.Any(), gomock.Any()).Return(nil)
			},
			statusCode: http.StatusCreated,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.CreateTask(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestTaskTransport_GetTasksByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskUsecase := mocks.NewMockTaskUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockTaskUsecase, cfg)

	tests := []struct {
		name       string
		userID     uuid.UUID
		mockFunc   func()
		statusCode int
	}{
		{
			name:   "Success",
			userID: uuid.New(),
			mockFunc: func() {
				tasks := []*dto.TaskDTO{
					{
						ID:          uuid.New(),
						Title:       "Test Task",
						Description: "Test Description",
						Status:      "todo",
						Importance:  1,
						Deadline:    time.Now().Add(24 * time.Hour),
					},
				}
				mockTaskUsecase.EXPECT().GetTasksByUserID(gomock.Any(), gomock.Any()).Return(tasks, nil)
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			req := httptest.NewRequest(http.MethodGet, "/tasks/user", nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, tt.userID.String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetTasksByUserID(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestTaskTransport_GetTasksByProjectID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskUsecase := mocks.NewMockTaskUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockTaskUsecase, cfg)

	router := mux.NewRouter()
	router.HandleFunc("/tasks/project/{projectId}", handler.GetTasksByProjectID).Methods("GET")

	tests := []struct {
		name       string
		projectID  uuid.UUID
		mockFunc   func()
		statusCode int
	}{
		{
			name:      "Success",
			projectID: uuid.New(),
			mockFunc: func() {
				tasks := []*dto.TaskDTO{
					{
						ID:          uuid.New(),
						Title:       "Test Task",
						Description: "Test Description",
						Status:      "todo",
						Importance:  1,
						Deadline:    time.Now().Add(24 * time.Hour),
					},
				}
				mockTaskUsecase.EXPECT().GetTasksByProjectID(gomock.Any(), gomock.Any()).Return(tasks, nil)
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			url := "/tasks/project/" + tt.projectID.String()
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestTaskTransport_UpdateTaskStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskUsecase := mocks.NewMockTaskUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockTaskUsecase, cfg)

	router := mux.NewRouter()
	router.HandleFunc("/tasks/{taskId}/status", handler.UpdateTaskStatus).Methods("PUT")

	tests := []struct {
		name       string
		taskID     uuid.UUID
		body       interface{}
		mockFunc   func()
		statusCode int
	}{
		{
			name:   "Success",
			taskID: uuid.New(),
			body:   map[string]string{},
			mockFunc: func() {
				mockTaskUsecase.EXPECT().UpdateTaskStatus(gomock.Any(), "waiting", gomock.Any(), gomock.Any()).Return(nil)
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			bodyBytes, _ := json.Marshal(tt.body)
			url := "/tasks/" + tt.taskID.String() + "/status?status=waiting"
			req := httptest.NewRequest("PUT", url, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestTaskTransport_DeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskUsecase := mocks.NewMockTaskUsecase(ctrl)
	cfg := &config.Config{}
	handler := New(mockTaskUsecase, cfg)

	router := mux.NewRouter()
	router.HandleFunc("/tasks/{taskId}", handler.DeleteTask).Methods("DELETE")

	tests := []struct {
		name       string
		taskID     uuid.UUID
		mockFunc   func()
		statusCode int
	}{
		{
			name:   "Success",
			taskID: uuid.New(),
			mockFunc: func() {
				mockTaskUsecase.EXPECT().DeleteTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			url := "/tasks/" + tt.taskID.String()
			req := httptest.NewRequest(http.MethodDelete, url, nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}
