package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	models "github.com/lzimin05/course-todo/internal/models/task"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/task"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/helpers"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	GetTasksByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Task, error)
	UpdateTask(ctx context.Context, title, description string, importance int, deadline time.Time, taskID, userID uuid.UUID) error
	UpdateTaskStatus(ctx context.Context, status string, taskID, userID uuid.UUID) error
	DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error
}

type TaskUsecase struct {
	repo TaskRepository
}

func New(repo TaskRepository) *TaskUsecase {
	return &TaskUsecase{repo: repo}
}

func (uc *TaskUsecase) CreateTask(ctx context.Context, req *dto.PostTaskDTO) error {
	const op = "TaskUseCase.CreateTask"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("title", req.Title)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return err
	}

	newTaskModel := &models.Task{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Importance:  req.Importance,
		Deadline:    req.Deadline,
		CreatedAt:   time.Now(),
		Status:      models.StatusWaiting,
	}

	_, err = uc.repo.CreateTask(ctx, newTaskModel)
	if err != nil {
		logger.WithError(err).Error("failed to create task")
		return err
	}

	return nil
}

func (uc *TaskUsecase) GetTasksByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.TaskDTO, error) {
	const op = "TaskUseCase.GetTaskByUserID"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("userID", userID)
	tasksmodel, err := uc.repo.GetTasksByUserID(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks by UserID")
		return nil, err
	}

	TasksDTO := make([]*dto.TaskDTO, len(tasksmodel))
	for i, taskmodel := range tasksmodel {
		TasksDTO[i] = &dto.TaskDTO{
			ID:          taskmodel.ID,
			UserID:      taskmodel.UserID,
			Title:       taskmodel.Title,
			Description: taskmodel.Description,
			Importance:  taskmodel.Importance,
			Deadline:    taskmodel.Deadline,
			Status:      taskmodel.Status,
			CreatedAt:   taskmodel.CreatedAt,
		}
	}

	return TasksDTO, nil
}

func (uc *TaskUsecase) UpdateTask(ctx context.Context, title, description string, importance int, deadline time.Time, taskID, userID uuid.UUID) error {
	const op = "TaskUseCase.UpdateTask"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("TaskID", taskID)
	err := uc.repo.UpdateTask(ctx, title, description, importance, deadline, taskID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to update task")
		return err
	}
	return nil
}

func (uc *TaskUsecase) UpdateTaskStatus(ctx context.Context, status string, taskID, userID uuid.UUID) error {
	const op = "TaskUseCase.UpdateTaskStatus"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("TaskID", taskID)
	err := uc.repo.UpdateTaskStatus(ctx, status, taskID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to update status for task")
		return err
	}
	return nil
}

func (uc *TaskUsecase) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	const op = "TaskUseCase.DeleteTask"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("TaskID", taskID)
	err := uc.repo.DeleteTask(ctx, taskID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to delete task")
		return err
	}
	return nil
}
