package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/internal/models/domains"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/task"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/task"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func TestTaskUsecase_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockTaskProjectRepository(ctrl)
	uc := New(mockTaskRepo, mockProjectRepo)

	userID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name          string
		request       *dto.PostTaskDTO
		setupContext  func() context.Context
		setupMocks    func()
		expectedError error
	}{
		{
			name: "successful task creation",
			request: &dto.PostTaskDTO{
				ProjectID:   projectID,
				Title:       "Test Task",
				Description: "Test Description",
				Importance:  1,
				Deadline:    time.Now().Add(24 * time.Hour),
			},
			setupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
				return logctx.WithLogger(ctx, logctx.NewLogger())
			},
			setupMocks: func() {
				mockProjectRepo.EXPECT().
					CheckProjectAccess(gomock.Any(), projectID, userID).
					Return(true, nil)

				mockTaskRepo.EXPECT().
					CreateTask(gomock.Any(), gomock.Any()).
					Return(&models.Task{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "no project access",
			request: &dto.PostTaskDTO{
				ProjectID:   projectID,
				Title:       "Test Task",
				Description: "Test Description",
				Importance:  1,
				Deadline:    time.Now().Add(24 * time.Hour),
			},
			setupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
				return logctx.WithLogger(ctx, logctx.NewLogger())
			},
			setupMocks: func() {
				mockProjectRepo.EXPECT().
					CheckProjectAccess(gomock.Any(), projectID, userID).
					Return(false, nil)
			},
			expectedError: errs.ErrNoAccess,
		},
		{
			name: "project access check error",
			request: &dto.PostTaskDTO{
				ProjectID:   projectID,
				Title:       "Test Task",
				Description: "Test Description",
				Importance:  1,
				Deadline:    time.Now().Add(24 * time.Hour),
			},
			setupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
				return logctx.WithLogger(ctx, logctx.NewLogger())
			},
			setupMocks: func() {
				mockProjectRepo.EXPECT().
					CheckProjectAccess(gomock.Any(), projectID, userID).
					Return(false, errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "create task repository error",
			request: &dto.PostTaskDTO{
				ProjectID:   projectID,
				Title:       "Test Task",
				Description: "Test Description",
				Importance:  1,
				Deadline:    time.Now().Add(24 * time.Hour),
			},
			setupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
				return logctx.WithLogger(ctx, logctx.NewLogger())
			},
			setupMocks: func() {
				mockProjectRepo.EXPECT().
					CheckProjectAccess(gomock.Any(), projectID, userID).
					Return(true, nil)

				mockTaskRepo.EXPECT().
					CreateTask(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			ctx := tt.setupContext()

			_, err := uc.CreateTask(ctx, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskUsecase_GetTasksByProjectID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockTaskProjectRepository(ctrl)
	uc := New(mockTaskRepo, mockProjectRepo)

	userID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name          string
		projectID     uuid.UUID
		setupContext  func() context.Context
		setupMocks    func()
		expectedTasks []*dto.TaskDTO
		expectedError error
	}{
		{
			name:      "successful get tasks by project",
			projectID: projectID,
			setupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
				return logctx.WithLogger(ctx, logctx.NewLogger())
			},
			setupMocks: func() {
				mockProjectRepo.EXPECT().
					CheckProjectAccess(gomock.Any(), projectID, userID).
					Return(true, nil)

				tasks := []*models.Task{
					{
						ID:          uuid.New(),
						ProjectID:   projectID,
						UserID:      userID,
						Title:       "Test Task 1",
						Description: "Test Description 1",
						Importance:  1,
						Deadline:    time.Now().Add(24 * time.Hour),
						CreatedAt:   time.Now(),
						Status:      models.StatusWaiting,
					},
				}

				mockTaskRepo.EXPECT().
					GetTasksByProjectID(gomock.Any(), projectID, userID).
					Return(tasks, nil)
			},
			expectedTasks: []*dto.TaskDTO{},
			expectedError: nil,
		},
		{
			name:      "no project access",
			projectID: projectID,
			setupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
				return logctx.WithLogger(ctx, logctx.NewLogger())
			},
			setupMocks: func() {
				mockProjectRepo.EXPECT().
					CheckProjectAccess(gomock.Any(), projectID, userID).
					Return(false, nil)
			},
			expectedTasks: nil,
			expectedError: errs.ErrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			ctx := tt.setupContext()

			tasks, err := uc.GetTasksByProjectID(ctx, tt.projectID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tasks)
			}
		})
	}
}

func TestTaskUsecase_UpdateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockTaskProjectRepository(ctrl)
	uc := New(mockTaskRepo, mockProjectRepo)

	tests := []struct {
		name          string
		title         string
		description   string
		importance    int
		deadline      time.Time
		taskID        uuid.UUID
		userID        uuid.UUID
		setupMocks    func()
		expectedError error
	}{
		{
			name:        "successful task update",
			title:       "Updated Task",
			description: "Updated Description",
			importance:  2,
			deadline:    time.Now().Add(48 * time.Hour),
			taskID:      uuid.New(),
			userID:      uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					UpdateTask(gomock.Any(), "Updated Task", "Updated Description", 2, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "repository error",
			title:       "Updated Task",
			description: "Updated Description",
			importance:  2,
			deadline:    time.Now().Add(48 * time.Hour),
			taskID:      uuid.New(),
			userID:      uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					UpdateTask(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())
			err := uc.UpdateTask(ctx, tt.title, tt.description, tt.importance, tt.deadline, tt.taskID, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskUsecase_GetTasksByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockTaskProjectRepository(ctrl)
	uc := New(mockTaskRepo, mockProjectRepo)

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMocks    func()
		expectedTasks []*models.Task
		expectedError error
	}{
		{
			name:   "successful get tasks",
			userID: uuid.New(),
			setupMocks: func() {
				tasks := []*models.Task{
					{
						ID:          uuid.New(),
						UserID:      uuid.New(),
						Title:       "Test Task 1",
						Description: "Test Description 1",
						Importance:  1,
						Deadline:    time.Now().Add(24 * time.Hour),
						CreatedAt:   time.Now(),
						Status:      models.StatusWaiting,
					},
					{
						ID:          uuid.New(),
						UserID:      uuid.New(),
						Title:       "Test Task 2",
						Description: "Test Description 2",
						Importance:  2,
						Deadline:    time.Now().Add(48 * time.Hour),
						CreatedAt:   time.Now(),
						Status:      models.StatusInProgress,
					},
				}

				mockTaskRepo.EXPECT().
					GetTasksByUserID(gomock.Any(), gomock.Any()).
					Return(tasks, nil)
			},
			expectedTasks: func() []*models.Task {
				return []*models.Task{
					{
						ID:          uuid.New(),
						UserID:      uuid.New(),
						Title:       "Test Task 1",
						Description: "Test Description 1",
						Importance:  1,
						Deadline:    time.Now().Add(24 * time.Hour),
						CreatedAt:   time.Now(),
						Status:      models.StatusWaiting,
					},
					{
						ID:          uuid.New(),
						UserID:      uuid.New(),
						Title:       "Test Task 2",
						Description: "Test Description 2",
						Importance:  2,
						Deadline:    time.Now().Add(48 * time.Hour),
						CreatedAt:   time.Now(),
						Status:      models.StatusInProgress,
					},
				}
			}(),
			expectedError: nil,
		},
		{
			name:   "repository error",
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					GetTasksByUserID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedTasks: nil,
			expectedError: errors.New("database error"),
		},
		{
			name:   "empty tasks list",
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					GetTasksByUserID(gomock.Any(), gomock.Any()).
					Return([]*models.Task{}, nil)
			},
			expectedTasks: []*models.Task{},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())
			tasks, err := uc.GetTasksByUserID(ctx, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, len(tt.expectedTasks))
			}
		})
	}
}

func TestTaskUsecase_UpdateTaskStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockTaskProjectRepository(ctrl)
	uc := New(mockTaskRepo, mockProjectRepo)

	tests := []struct {
		name          string
		status        string
		taskID        uuid.UUID
		userID        uuid.UUID
		setupMocks    func()
		expectedError error
	}{
		{
			name:   "successful status update",
			status: models.StatusCompleted,
			taskID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), models.StatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			status: models.StatusCompleted,
			taskID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), models.StatusCompleted, gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:   "invalid status",
			status: "invalid_status",
			taskID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					UpdateTaskStatus(gomock.Any(), "invalid_status", gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil, // Пока нет валидации в usecase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())
			err := uc.UpdateTaskStatus(ctx, tt.status, tt.taskID, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskUsecase_DeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockTaskProjectRepository(ctrl)
	uc := New(mockTaskRepo, mockProjectRepo)

	tests := []struct {
		name          string
		taskID        uuid.UUID
		userID        uuid.UUID
		setupMocks    func()
		expectedError error
	}{
		{
			name:   "successful delete",
			taskID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					DeleteTask(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "task not found",
			taskID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					DeleteTask(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errs.ErrTaskNotFound)
			},
			expectedError: errs.ErrTaskNotFound,
		},
		{
			name:   "repository error",
			taskID: uuid.New(),
			userID: uuid.New(),
			setupMocks: func() {
				mockTaskRepo.EXPECT().
					DeleteTask(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())
			err := uc.DeleteTask(ctx, tt.taskID, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew_TaskUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockTaskProjectRepository(ctrl)

	uc := New(mockTaskRepo, mockProjectRepo)

	assert.NotNil(t, uc)
	assert.Equal(t, mockTaskRepo, uc.repo)
	assert.Equal(t, mockProjectRepo, uc.projectRepo)
}
