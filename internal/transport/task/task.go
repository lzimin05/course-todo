package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/models/domains"
	models "github.com/lzimin05/course-todo/internal/models/task"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/task"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/transport/utils/handler"
	response "github.com/lzimin05/course-todo/internal/transport/utils/response"
	validation "github.com/lzimin05/course-todo/internal/transport/utils/validation/task"
)

//go:generate mockgen -source=task.go -destination=../../usecase/mocks/task_usecase_mock.go -package=mocks TaskUsecase
type TaskUsecase interface {
	CreateTask(ctx context.Context, req *dto.PostTaskDTO) error
	GetTasksByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.TaskDTO, error)
	GetTasksByProjectID(ctx context.Context, projectID uuid.UUID) ([]*dto.TaskDTO, error)
	UpdateTask(ctx context.Context, title, description string, importance int, deadline time.Time, taskID, userID uuid.UUID) error
	UpdateTaskStatus(ctx context.Context, status string, taskID, userID uuid.UUID) error
	DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error
}

type TaskHandler struct {
	uc     TaskUsecase
	config *config.Config
}

func New(uc TaskUsecase, cfg *config.Config) *TaskHandler {
	return &TaskHandler{
		uc:     uc,
		config: cfg,
	}
}

// CreateTask создает новую задачу
// @Summary      Создать новую задачу
// @Description  Создает новую задачу для текущего пользователя
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body  dto.PostTaskDTO  true  "Данные для создания задачи"
// @Success      201  "Задача создана"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Нет доступа к проекту"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /todo/create [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	const op = "TaskHandler.CreateTask"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	var req dto.PostTaskDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode task data")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := validation.ValidationTask(req.Title, req.Importance, req.Deadline); err != nil {
		logger.Warn("validation failed: ", err.Error())
		response.SendError(r.Context(), w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.uc.CreateTask(r.Context(), &req)
	if err != nil {
		logger.WithError(err).Error("failed to create task")
		handler.HandleError(r.Context(), w, err, "Failed to create task")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, nil)
}

// GetTasksByUserID получает все задачи пользователя
// @Summary      Получить задачи пользователя
// @Description  Возвращает список всех задач текущего пользователя
// @Tags         tasks
// @Produce      json
// @Success      200  {array}  dto.TaskDTO "Список задач"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /todo/all [get]
func (h *TaskHandler) GetTasksByUserID(w http.ResponseWriter, r *http.Request) {
	const op = "TaskHandler.GetTasksByUserID"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)
	userIDValue, ok := r.Context().Value(domains.UserIDKey{}).(string)
	if !ok {
		logger.Warn("userID not found in context")
		response.SendError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		logger.Warnf("invalid userID type in context, got %T, expected uuid.UUID", userIDValue)
		response.SendError(r.Context(), w, http.StatusBadRequest, "invalid userID format in context")
		return
	}
	tasks, err := h.uc.GetTasksByUserID(r.Context(), userID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "failed to get tasks")
		return
	}
	response.SendJSONResponse(r.Context(), w, http.StatusOK, tasks)
}

// GetTasksByProjectID получает все задачи проекта
// @Summary      Получить задачи проекта
// @Description  Возвращает список всех задач указанного проекта
// @Tags         tasks
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Success      200  {array}  dto.TaskDTO "Список задач"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Нет доступа к проекту"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId}/tasks [get]
func (h *TaskHandler) GetTasksByProjectID(w http.ResponseWriter, r *http.Request) {
	const op = "TaskHandler.GetTasksByProjectID"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	tasks, err := h.uc.GetTasksByProjectID(r.Context(), projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks by project")
		handler.HandleError(r.Context(), w, err, "Failed to get tasks")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, tasks)
}

// UpdateTask обновляет задачу
// @Summary      Обновить задачу
// @Description  Обновляет существующую задачу пользователя
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        taskId  path  string  true  "ID задачи"
// @Param        task    body  dto.PostTaskDTO  true  "Данные для обновления задачи"
// @Success      200  "Задача обновлена"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /todo/{taskId}/edit [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	const op = "TaskHandler.UpdateTask"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)
	userIDValue, ok := r.Context().Value(domains.UserIDKey{}).(string)
	if !ok {
		logger.Warn("userID not found in context")
		response.SendError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		logger.Warnf("invalid userID type in context, got %T, expected uuid.UUID", userIDValue)
		response.SendError(r.Context(), w, http.StatusBadRequest, "invalid userID format in context")
		return
	}

	taskID, err := uuid.Parse(mux.Vars(r)["taskId"])
	if err != nil {
		logger.Errorf("failed to parse taskID: %v", err)
		http.Error(w, "failed to parse taskID", http.StatusBadRequest)
		return
	}
	var req dto.PostTaskDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode task")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invailid request")
		return
	}

	err = validation.ValidationTask(req.Title, req.Importance, req.Deadline)
	if err != nil {
		logger.Warn("valitationTask error: ", err.Error())
		response.SendError(r.Context(), w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.uc.UpdateTask(r.Context(), req.Title, req.Description, req.Importance, req.Deadline, taskID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to update task")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "failed to update task")
		return
	}
	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}

// UpdateTaskStatus обновляет статус задачи
// @Summary      Обновить статус задачи
// @Description  Обновляет статус существующей задачи пользователя
// @Tags         tasks
// @Produce      json
// @Param        taskId  path   string  true  "ID задачи"
// @Param        status  query  string  true  "Новый статус задачи (waiting, in_progress, completed)"
// @Success      200  "Статус задачи обновлен"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /todo/{taskId}/edit [patch]
func (h *TaskHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	const op = "TaskHandler.UpdateTaskStatus"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)
	userIDValue, ok := r.Context().Value(domains.UserIDKey{}).(string)
	if !ok {
		logger.Warn("userID not found in context")
		response.SendError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		logger.Warnf("invalid userID type in context, got %T, expected uuid.UUID", userIDValue)
		response.SendError(r.Context(), w, http.StatusBadRequest, "invalid userID format in context")
		return
	}

	taskID, err := uuid.Parse(mux.Vars(r)["taskId"])
	if err != nil {
		logger.Errorf("failed to parse taskID: %v", err)
		http.Error(w, "failed to parse taskID", http.StatusBadRequest)
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Status parameter is required")
		return
	}
	validStatuses := map[string]bool{
		models.StatusWaiting:    true,
		models.StatusInProgress: true,
		models.StatusCompleted:  true,
	}
	if !validStatuses[status] {
		response.SendError(r.Context(), w, http.StatusBadRequest,
			fmt.Sprintf("Invalid status. Allowed values: %s, %s, %s", models.StatusWaiting, models.StatusInProgress, models.StatusCompleted))
		return
	}
	err = h.uc.UpdateTaskStatus(r.Context(), status, taskID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to update status for task")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "failed to update status for task")
		return
	}
	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}

// DeleteTask удаляет задачу
// @Summary      Удалить задачу
// @Description  Удаляет существующую задачу пользователя
// @Tags         tasks
// @Produce      json
// @Param        taskId  path  string  true  "ID задачи"
// @Success      200  "Задача удалена"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /todo/{taskId} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	const op = "TaskHandler.DeleteTask"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)
	userIDValue, ok := r.Context().Value(domains.UserIDKey{}).(string)
	if !ok {
		logger.Warn("userID not found in context")
		response.SendError(r.Context(), w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		logger.Warnf("invalid userID type in context, got %T, expected uuid.UUID", userIDValue)
		response.SendError(r.Context(), w, http.StatusBadRequest, "invalid userID format in context")
		return
	}

	taskID, err := uuid.Parse(mux.Vars(r)["taskId"])
	if err != nil {
		logger.Errorf("failed to parse taskID: %v", err)
		http.Error(w, "failed to parse taskID", http.StatusBadRequest)
		return
	}
	err = h.uc.DeleteTask(r.Context(), taskID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to delete task")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}
