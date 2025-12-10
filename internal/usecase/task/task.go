package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/task"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/task"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/helpers"
)

//go:generate mockgen -source=task.go -destination=../mocks/task_mocks.go -package=mocks TaskRepository,TaskProjectRepository
type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	GetTasksByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Task, error)
	GetTasksByProjectID(ctx context.Context, projectID, userID uuid.UUID) ([]*models.Task, error)
	UpdateTask(ctx context.Context, title, description string, importance int, deadline time.Time, taskID, userID uuid.UUID) error
	UpdateTaskStatus(ctx context.Context, status string, taskID, userID uuid.UUID) error
	DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error
}

type TaskProjectRepository interface {
	CheckProjectAccess(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
}

type TaskUsecase struct {
	repo        TaskRepository
	projectRepo TaskProjectRepository
}

func New(repo TaskRepository, projectRepo TaskProjectRepository) *TaskUsecase {
	return &TaskUsecase{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

func (uc *TaskUsecase) CreateTask(ctx context.Context, req *dto.PostTaskDTO) (*dto.CreateTaskDTO, error) {
	const op = "TaskUseCase.CreateTask"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("title", req.Title)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	// Проверяем права доступа к проекту
	hasAccess, err := uc.projectRepo.CheckProjectAccess(ctx, req.ProjectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to check project access")
		return nil, err
	}

	if !hasAccess {
		logger.Warn("user doesn't have access to project")
		return nil, errs.ErrNoAccess
	}

	newTaskModel := &models.Task{
		ID:          uuid.New(),
		ProjectID:   req.ProjectID,
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
		return nil, err
	}

	return &dto.CreateTaskDTO{
		ID: newTaskModel.ID,
	}, nil
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
			ProjectID:   taskmodel.ProjectID,
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

func (uc *TaskUsecase) GetTasksByProjectID(ctx context.Context, projectID uuid.UUID) ([]*dto.TaskDTO, error) {
	const op = "TaskUseCase.GetTasksByProjectID"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	// Проверяем права доступа к проекту
	hasAccess, err := uc.projectRepo.CheckProjectAccess(ctx, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to check project access")
		return nil, err
	}

	if !hasAccess {
		logger.Warn("user doesn't have access to project")
		return nil, errs.ErrNoAccess
	}

	tasksmodel, err := uc.repo.GetTasksByProjectID(ctx, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks by ProjectID")
		return nil, err
	}

	TasksDTO := make([]*dto.TaskDTO, len(tasksmodel))
	for i, taskmodel := range tasksmodel {
		TasksDTO[i] = &dto.TaskDTO{
			ID:          taskmodel.ID,
			ProjectID:   taskmodel.ProjectID,
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
