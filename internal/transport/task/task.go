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
	response "github.com/lzimin05/course-todo/internal/transport/utils/response"
)

type TaskUsecase interface {
	CreateTask(ctx context.Context, userID uuid.UUID, title, description string, importance int, deadline, createdAt time.Time, status string) (*models.Task, error)
	GetTasksByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Task, error)
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

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	const op = "TaskHandler.CreateTask"
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

	var req dto.TaskDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode task")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invailid request")
		return
	}

	if req.Title == "" {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Title is required")
		return
	}

	if req.Importance < 1 || req.Importance > 3 {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Importance must be between 1 and 3")
		return
	}

	if req.Deadline.Before(time.Now()) {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Deadline must be in the future")
		return
	}

	task, err := h.uc.CreateTask(r.Context(), userID, req.Title, req.Description, req.Importance, req.Deadline, time.Now().Truncate(time.Second), models.StatusWaiting)
	if err != nil {
		logger.WithError(err).Error("failed to create task")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to create task")
		return
	}
	response.SendJSONResponse(r.Context(), w, http.StatusOK, task)
}

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
	var req dto.TaskDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode task")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invailid request")
		return
	}
	if req.Title == "" {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Title is required")
		return
	}

	if req.Importance < 1 || req.Importance > 3 {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Importance must be between 1 and 3")
		return
	}

	if req.Deadline.Before(time.Now()) {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Deadline must be in the future")
		return
	}
	err = h.uc.UpdateTask(r.Context(), req.Title, req.Description, req.Importance, req.Deadline, taskID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to update task")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "failed to update task")
		return
	}
	response.SendJSONResponse(r.Context(), w, http.StatusOK, map[string]string{"status": "Task updated successfully"})
}

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
	response.SendJSONResponse(r.Context(), w, http.StatusOK, map[string]string{"status": "Status task updated successfully"})
}

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
	response.SendJSONResponse(r.Context(), w, http.StatusOK, map[string]string{"status": "Status task delete successfully"})
}
