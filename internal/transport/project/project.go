package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lzimin05/course-todo/config"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/project"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/transport/utils/handler"
	response "github.com/lzimin05/course-todo/internal/transport/utils/response"
)

type ProjectUsecase interface {
	CreateProject(ctx context.Context, req *dto.PostProjectDTO) (*dto.ProjectDTO, error)
	GetUserProjects(ctx context.Context) ([]*dto.ProjectDTO, error)
	GetProjectByID(ctx context.Context, projectID uuid.UUID) (*dto.ProjectDTO, error)
	AddProjectMember(ctx context.Context, projectID uuid.UUID, req *dto.AddMemberDTO) error
	GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*dto.ProjectMemberDTO, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID) error
	RemoveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID) error
	UpdateProject(ctx context.Context, projectID uuid.UUID, req *dto.UpdateProjectDTO) (*dto.ProjectDTO, error)
	LeaveProject(ctx context.Context, projectID uuid.UUID) error
}

type ProjectHandler struct {
	uc     ProjectUsecase
	config *config.Config
}

func New(uc ProjectUsecase, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{
		uc:     uc,
		config: cfg,
	}
}

// handleError обрабатывает ошибки и возвращает соответствующий HTTP статус
// CreateProject создает новый проект
// @Summary      Создать новый проект
// @Description  Создает новый проект для текущего пользователя
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project  body  dto.PostProjectDTO  true  "Данные для создания проекта"
// @Success      201  {object} dto.ProjectDTO "Проект создан"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.CreateProject"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	var req dto.PostProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode project")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.Name == "" {
		logger.Warn("empty project name")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Project name is required")
		return
	}

	project, err := h.uc.CreateProject(r.Context(), &req)
	if err != nil {
		logger.WithError(err).Error("failed to create project")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, project)
}

// GetUserProjects получает все проекты пользователя
// @Summary      Получить проекты пользователя
// @Description  Возвращает список всех проектов текущего пользователя
// @Tags         projects
// @Produce      json
// @Success      200  {array}  dto.ProjectDTO "Список проектов"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects [get]
func (h *ProjectHandler) GetUserProjects(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.GetUserProjects"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projects, err := h.uc.GetUserProjects(r.Context())
	if err != nil {
		logger.WithError(err).Error("failed to get projects")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to get projects")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, projects)
}

// GetProjectByID получает проект по ID
// @Summary      Получить проект по ID
// @Description  Возвращает информацию о проекте по его ID
// @Tags         projects
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Success      200  {object} dto.ProjectDTO "Информация о проекте"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Нет доступа к проекту"
// @Failure      404  {object} dto.ErrorResponse "Проект не найден"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId} [get]
func (h *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.GetProjectByID"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	project, err := h.uc.GetProjectByID(r.Context(), projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project")
		handler.HandleError(r.Context(), w, err, "Failed to get project")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, project)
}

// AddProjectMember добавляет участника в проект
// @Summary      Добавить участника в проект
// @Description  Добавляет пользователя в проект в качестве участника
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Param        member     body  dto.AddMemberDTO  true  "Данные участника"
// @Success      200  "Участник добавлен"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId}/members [post]
func (h *ProjectHandler) AddProjectMember(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.AddProjectMember"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req dto.AddMemberDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode member data")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	err = h.uc.AddProjectMember(r.Context(), projectID, &req)
	if err != nil {
		logger.WithError(err).Error("failed to add project member")
		handler.HandleError(r.Context(), w, err, "Failed to add member")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}

// GetProjectMembers получает список участников проекта
// @Summary      Получить участников проекта
// @Description  Возвращает список всех участников проекта
// @Tags         projects
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Success      200  {array}  dto.ProjectMemberDTO "Список участников"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Нет доступа к проекту"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId}/members [get]
func (h *ProjectHandler) GetProjectMembers(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.GetProjectMembers"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	members, err := h.uc.GetProjectMembers(r.Context(), projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project members")
		handler.HandleError(r.Context(), w, err, "Failed to get members")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, members)
}

// DeleteProject удаляет проект
// @Summary      Удалить проект
// @Description  Удаляет проект (только владелец)
// @Tags         projects
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Success      200  "Проект удален"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404  {object} dto.ErrorResponse "Проект не найден"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId} [delete]
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.DeleteProject"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	err = h.uc.DeleteProject(r.Context(), projectID)
	if err != nil {
		logger.WithError(err).Error("failed to delete project")
		handler.HandleError(r.Context(), w, err, "Failed to delete project")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}

// RemoveProjectMember удаляет участника из проекта
// @Summary      Удалить участника из проекта
// @Description  Удаляет участника из проекта (только владелец)
// @Tags         projects
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Param        userId     path  string  true  "ID пользователя"
// @Success      200  "Участник удален"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404  {object} dto.ErrorResponse "Участник не найден"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId}/members/{userId} [delete]
func (h *ProjectHandler) RemoveProjectMember(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.RemoveProjectMember"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	vars := mux.Vars(r)

	projectID, err := uuid.Parse(vars["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		logger.WithError(err).Warn("invalid user ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.uc.RemoveProjectMember(r.Context(), projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to remove project member")
		handler.HandleError(r.Context(), w, err, "Failed to remove member")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}

// UpdateProject обновляет название и описание проекта
// @Summary      Обновить проект
// @Description  Обновляет название и описание проекта (только владелец)
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Param        project    body  dto.UpdateProjectDTO  true  "Данные для обновления проекта"
// @Success      200  {object} dto.ProjectDTO "Обновленный проект"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404  {object} dto.ErrorResponse "Проект не найден"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId} [put]
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.UpdateProject"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req dto.UpdateProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode project update data")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.Name == "" {
		logger.Warn("empty project name")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Project name is required")
		return
	}

	project, err := h.uc.UpdateProject(r.Context(), projectID, &req)
	if err != nil {
		logger.WithError(err).Error("failed to update project")
		handler.HandleError(r.Context(), w, err, "Failed to update project")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, project)
}

// LeaveProject позволяет пользователю покинуть проект
// @Summary      Покинуть проект
// @Description  Позволяет участнику покинуть проект (владелец не может покинуть проект)
// @Tags         projects
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Success      200  "Пользователь покинул проект"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Недостаточно прав или владелец не может покинуть проект"
// @Failure      404  {object} dto.ErrorResponse "Проект не найден"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId}/leave [post]
func (h *ProjectHandler) LeaveProject(w http.ResponseWriter, r *http.Request) {
	const op = "ProjectHandler.LeaveProject"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	err = h.uc.LeaveProject(r.Context(), projectID)
	if err != nil {
		logger.WithError(err).Error("failed to leave project")
		handler.HandleError(r.Context(), w, err, "Failed to leave project")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}
