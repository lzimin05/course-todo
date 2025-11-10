package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	models "github.com/lzimin05/course-todo/internal/models/task"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

type TaskRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

const (
	CreateTaskQuery = `INSERT INTO todo.task (id, project_id, user_id, title, description, importance, status, created_at, deadline)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, project_id, user_id, title, description, importance, status, created_at, deadline`

	GetTasksByProjectIDQuery = `SELECT t.id, t.project_id, t.user_id, t.title, t.description, t.importance, t.status, t.created_at, t.deadline
	FROM todo.task t
	JOIN todo.project_member pm ON t.project_id = pm.project_id
	WHERE t.project_id = $1 AND pm.user_id = $2`

	GetTasksByUserIDQuery = `SELECT t.id, t.project_id, t.user_id, t.title, t.description, t.importance, t.status, t.created_at, t.deadline
	FROM todo.task t
	JOIN todo.project_member pm ON t.project_id = pm.project_id
	WHERE pm.user_id = $1`

	UpdateTaskQuery = `UPDATE todo.task SET title = $1, description = $2, importance = $3, deadline = $4
	WHERE id = $5 AND project_id IN (
		SELECT pm.project_id FROM todo.project_member pm WHERE pm.user_id = $6
	)`

	UpdateTaskStatusQuery = `UPDATE todo.task SET status = $1 
	WHERE id = $2 AND project_id IN (
		SELECT pm.project_id FROM todo.project_member pm WHERE pm.user_id = $3
	)`

	DeleteTaskQuery = `DELETE FROM todo.task
	WHERE id = $1 AND project_id IN (
		SELECT pm.project_id FROM todo.project_member pm WHERE pm.user_id = $2
	)`

	TaskExistenceForUserQuery = `SELECT EXISTS(
		SELECT 1 FROM todo.task t
		JOIN todo.project_member pm ON t.project_id = pm.project_id
		WHERE t.id = $1 AND pm.user_id = $2
	)`
)

func (r *TaskRepository) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	const op = "TaskRepository.CreateTask"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("title", task.Title)

	err := r.db.QueryRowContext(ctx, CreateTaskQuery,
		task.ID, task.ProjectID, task.UserID, task.Title, task.Description, task.Importance, task.Status, task.CreatedAt, task.Deadline).
		Scan(&task.ID, &task.ProjectID, &task.UserID, &task.Title, &task.Description, &task.Importance, &task.Status, &task.CreatedAt, &task.Deadline)
	if err != nil {
		logger.WithError(err).Warn("failed to create task")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return task, nil
}

func (r *TaskRepository) GetTasksByProjectID(ctx context.Context, projectID, userID uuid.UUID) ([]*models.Task, error) {
	const op = "TaskRepository.GetTasksByProjectID"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("ProjectID", projectID)

	rows, err := r.db.QueryContext(ctx, GetTasksByProjectIDQuery, projectID, userID)
	if err != nil {
		logger.WithError(err).Warn("failed to get tasks by project")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.ProjectID, &t.UserID, &t.Title, &t.Description, &t.Importance, &t.Status, &t.CreatedAt, &t.Deadline)
		if err != nil {
			logger.WithError(err).Warn("failed to scan task")
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *TaskRepository) GetTasksByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Task, error) {
	const op = "TaskRepository.GetByUserID"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("UserID", userID)
	rows, err := r.db.QueryContext(ctx, GetTasksByUserIDQuery, userID)
	if err != nil {
		logger.WithError(err).Warn("failed to get tasks")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.ProjectID, &t.UserID, &t.Title, &t.Description, &t.Importance, &t.Status, &t.CreatedAt, &t.Deadline)
		if err != nil {
			logger.WithError(err).Warn("failed to get task in loop")
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, title, description string, importance int, deadline time.Time, taskID, userID uuid.UUID) error {
	const op = "TaskRepository.UpdateTask"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("TaskID", taskID)
	_, err := r.db.ExecContext(ctx, UpdateTaskQuery, title, description, importance, deadline, taskID, userID)
	if err != nil {
		logger.WithError(err).Warn("failed to update task")
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, status string, taskID, userID uuid.UUID) error {
	const op = "TaskRepository.UpdateTaskStatus"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("TaskID", taskID)
	_, err := r.db.ExecContext(ctx, UpdateTaskStatusQuery, status, taskID, userID)
	if err != nil {
		logger.WithError(err).Warn("failed to update status for task")
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	const op = "TaskRepository.DeleteTask"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("TaskID", taskID)

	var exists bool
	err := r.db.QueryRowContext(ctx, TaskExistenceForUserQuery, taskID, userID).Scan(&exists)

	if err != nil {
		logger.WithError(err).Error("failed to check task existence")
		return fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		logger.Warn("task not found or doesn't belong to user")
		return fmt.Errorf("%s: %w", op, sql.ErrNoRows)
	}

	_, err = r.db.ExecContext(ctx, DeleteTaskQuery, taskID, userID)
	if err != nil {
		logger.WithError(err).Warn("failed to delete task")
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
