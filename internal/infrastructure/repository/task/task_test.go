package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	models "github.com/lzimin05/course-todo/internal/models/task"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func TestTaskRepository_CreateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	taskID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()
	deadline := time.Now().Add(24 * time.Hour)

	task := &models.Task{
		ID:          taskID,
		ProjectID:   projectID,
		UserID:      userID,
		Title:       "Test Task",
		Description: "Test Description",
		Importance:  1,
		Status:      "pending",
		CreatedAt:   createdAt,
		Deadline:    deadline,
	}

	tests := []struct {
		name        string
		task        *models.Task
		setupMocks  func()
		expectedErr bool
	}{
		{
			name: "successful task creation",
			task: task,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "title", "description", "importance", "status", "created_at", "deadline"}).
					AddRow(taskID, projectID, userID, "Test Task", "Test Description", 1, "pending", createdAt, deadline)

				mock.ExpectQuery(`INSERT INTO todo.task`).
					WithArgs(taskID, projectID, userID, "Test Task", "Test Description", 1, "pending", createdAt, deadline).
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name: "database error",
			task: task,
			setupMocks: func() {
				mock.ExpectQuery(`INSERT INTO todo.task`).
					WithArgs(taskID, projectID, userID, "Test Task", "Test Description", 1, "pending", createdAt, deadline).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			result, err := repo.CreateTask(ctx, tt.task)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "TaskRepository.CreateTask")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.task.ID, result.ID)
				assert.Equal(t, tt.task.Title, result.Title)
				assert.Equal(t, tt.task.Description, result.Description)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskRepository_GetTasksByProjectID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	projectID := uuid.New()
	userID := uuid.New()
	taskID := uuid.New()
	createdAt := time.Now()
	deadline := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		projectID   uuid.UUID
		userID      uuid.UUID
		setupMocks  func()
		expectedErr bool
		expectTasks int
	}{
		{
			name:      "successful tasks retrieval",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "title", "description", "importance", "status", "created_at", "deadline"}).
					AddRow(taskID, projectID, userID, "Task 1", "Description 1", 1, "pending", createdAt, deadline).
					AddRow(uuid.New(), projectID, userID, "Task 2", "Description 2", 2, "completed", createdAt, deadline)

				mock.ExpectQuery(`SELECT t.id, t.project_id, t.user_id, t.title, t.description, t.importance, t.status, t.created_at, t.deadline`).
					WithArgs(projectID, userID).
					WillReturnRows(rows)
			},
			expectedErr: false,
			expectTasks: 2,
		},
		{
			name:      "no tasks found",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "title", "description", "importance", "status", "created_at", "deadline"})

				mock.ExpectQuery(`SELECT t.id, t.project_id, t.user_id, t.title, t.description, t.importance, t.status, t.created_at, t.deadline`).
					WithArgs(projectID, userID).
					WillReturnRows(rows)
			},
			expectedErr: false,
			expectTasks: 0,
		},
		{
			name:      "database error",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT t.id, t.project_id, t.user_id, t.title, t.description, t.importance, t.status, t.created_at, t.deadline`).
					WithArgs(projectID, userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
			expectTasks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			tasks, err := repo.GetTasksByProjectID(ctx, tt.projectID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "TaskRepository.GetTasksByProjectID")
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, tt.expectTasks)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskRepository_GetTasksByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()
	createdAt := time.Now()
	deadline := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		userID      uuid.UUID
		setupMocks  func()
		expectedErr bool
		expectTasks int
	}{
		{
			name:   "successful tasks retrieval by user",
			userID: userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "title", "description", "importance", "status", "created_at", "deadline"}).
					AddRow(taskID, projectID, userID, "User Task 1", "Description 1", 1, "pending", createdAt, deadline)

				mock.ExpectQuery(`SELECT t.id, t.project_id, t.user_id, t.title, t.description, t.importance, t.status, t.created_at, t.deadline`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedErr: false,
			expectTasks: 1,
		},
		{
			name:   "database error",
			userID: userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT t.id, t.project_id, t.user_id, t.title, t.description, t.importance, t.status, t.created_at, t.deadline`).
					WithArgs(userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
			expectTasks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			tasks, err := repo.GetTasksByUserID(ctx, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "TaskRepository.GetByUserID")
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, tt.expectTasks)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskRepository_UpdateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	taskID := uuid.New()
	userID := uuid.New()
	deadline := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		title       string
		description string
		importance  int
		deadline    time.Time
		taskID      uuid.UUID
		userID      uuid.UUID
		setupMocks  func()
		expectedErr bool
	}{
		{
			name:        "successful task update",
			title:       "Updated Task",
			description: "Updated Description",
			importance:  2,
			deadline:    deadline,
			taskID:      taskID,
			userID:      userID,
			setupMocks: func() {
				mock.ExpectExec(`UPDATE todo.task SET title = \$1, description = \$2, importance = \$3, deadline = \$4`).
					WithArgs("Updated Task", "Updated Description", 2, deadline, taskID, userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: false,
		},
		{
			name:        "database error",
			title:       "Updated Task",
			description: "Updated Description",
			importance:  2,
			deadline:    deadline,
			taskID:      taskID,
			userID:      userID,
			setupMocks: func() {
				mock.ExpectExec(`UPDATE todo.task SET title = \$1, description = \$2, importance = \$3, deadline = \$4`).
					WithArgs("Updated Task", "Updated Description", 2, deadline, taskID, userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := repo.UpdateTask(ctx, tt.title, tt.description, tt.importance, tt.deadline, tt.taskID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "TaskRepository.UpdateTask")
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskRepository_UpdateTaskStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	taskID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		status      string
		taskID      uuid.UUID
		userID      uuid.UUID
		setupMocks  func()
		expectedErr bool
	}{
		{
			name:   "successful status update",
			status: "completed",
			taskID: taskID,
			userID: userID,
			setupMocks: func() {
				mock.ExpectExec(`UPDATE todo.task SET status = \$1`).
					WithArgs("completed", taskID, userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: false,
		},
		{
			name:   "database error",
			status: "completed",
			taskID: taskID,
			userID: userID,
			setupMocks: func() {
				mock.ExpectExec(`UPDATE todo.task SET status = \$1`).
					WithArgs("completed", taskID, userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := repo.UpdateTaskStatus(ctx, tt.status, tt.taskID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "TaskRepository.UpdateTaskStatus")
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskRepository_DeleteTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	taskID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		taskID      uuid.UUID
		userID      uuid.UUID
		setupMocks  func()
		expectedErr bool
	}{
		{
			name:   "successful task deletion",
			taskID: taskID,
			userID: userID,
			setupMocks: func() {
				// Mock existence check
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(taskID, userID).
					WillReturnRows(rows)

				// Mock deletion
				mock.ExpectExec(`DELETE FROM todo.task`).
					WithArgs(taskID, userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: false,
		},
		{
			name:   "task not found",
			taskID: taskID,
			userID: userID,
			setupMocks: func() {
				// Mock existence check - task doesn't exist
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(taskID, userID).
					WillReturnRows(rows)
			},
			expectedErr: true,
		},
		{
			name:   "existence check error",
			taskID: taskID,
			userID: userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(taskID, userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := repo.DeleteTask(ctx, tt.taskID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "TaskRepository.DeleteTask")
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNew_TaskRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
