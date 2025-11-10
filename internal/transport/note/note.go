package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lzimin05/course-todo/config"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/note"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/transport/utils/handler"
	"github.com/lzimin05/course-todo/internal/transport/utils/response"
)

type INoteUsecase interface {
	GetAllNotes(ctx context.Context) ([]*dto.NoteDTO, error)
	GetNotesByProject(ctx context.Context, projectID uuid.UUID) ([]*dto.NoteDTO, error)
	CreateNote(ctx context.Context, req dto.CreateOrUpdateNote) (*dto.CreateNoteDTO, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, req dto.CreateOrUpdateNote) error
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
}

type NoteHandler struct {
	uc     INoteUsecase
	config *config.Config
}

func NewNoteHandler(uc INoteUsecase, conf *config.Config) *NoteHandler {
	return &NoteHandler{
		uc:     uc,
		config: conf,
	}
}

// GetAllNotes получает все заметки
// @Summary      Получить все заметки
// @Description  Возвращает список всех заметок пользователя
// @Tags         notes
// @Produce      json
// @Success      200  {array}  dto.NoteDTO "Список заметок"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /notes/all [get]
func (h *NoteHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	const op = "NoteHandler.GetAllNotes"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	notes, err := h.uc.GetAllNotes(r.Context())
	if err != nil {
		logger.WithError(err).Error("failed to get notes")
		handler.HandleError(r.Context(), w, err, "Internal server error")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, notes)
}

// GetNotesByProject получает все заметки проекта
// @Summary      Получить заметки проекта
// @Description  Возвращает список всех заметок указанного проекта
// @Tags         notes
// @Produce      json
// @Param        projectId  path  string  true  "ID проекта"
// @Success      200  {array}  dto.NoteDTO "Список заметок"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      403  {object} dto.ErrorResponse "Нет доступа к проекту"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /projects/{projectId}/notes [get]
func (h *NoteHandler) GetNotesByProject(w http.ResponseWriter, r *http.Request) {
	const op = "NoteHandler.GetNotesByProject"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	projectID, err := uuid.Parse(mux.Vars(r)["projectId"])
	if err != nil {
		logger.WithError(err).Warn("invalid project ID")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	notes, err := h.uc.GetNotesByProject(r.Context(), projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get notes by project")
		handler.HandleError(r.Context(), w, err, "Failed to get notes")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, notes)
}

// CreateNote создает новую заметку
// @Summary      Создать новую заметку
// @Description  Создает новую заметку для пользователя
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        note  body  dto.CreateOrUpdateNote  true  "Данные для создания заметки"
// @Success      201  {object} dto.CreateNoteDTO "ID созданной заметки"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /notes/create [post]
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	const op = "NoteHandler.CreateNote"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	var req dto.CreateOrUpdateNote
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode request")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request format")
		return
	}

	noteID, err := h.uc.CreateNote(r.Context(), req)
	if err != nil {
		logger.WithError(err).Error("failed to create note")
		handler.HandleError(r.Context(), w, err, "Failed to create note")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, noteID)
}

// UpdateNote обновляет заметку
// @Summary      Обновить заметку
// @Description  Обновляет существующую заметку пользователя
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        noteId  path  string  true  "ID заметки"
// @Param        note    body  dto.CreateOrUpdateNote  true  "Данные для обновления заметки"
// @Success      204  "Заметка успешно обновлена"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      404  {object} dto.ErrorResponse "Заметка не найдена"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /notes/{noteId}/edit [put]
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	const op = "NoteHandler.UpdateNote"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	noteID, err := uuid.Parse(mux.Vars(r)["noteId"])
	if err != nil {
		logger.WithError(err).Warn("invalid note ID format")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid note ID format")
		return
	}

	var req dto.CreateOrUpdateNote
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode request")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request format")
		return
	}

	err = h.uc.UpdateNote(r.Context(), noteID, req)
	if err != nil {
		logger.WithError(err).Error("failed to update note")
		handler.HandleError(r.Context(), w, err, "Failed to update note")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteNote удаляет заметку
// @Summary      Удалить заметку
// @Description  Удаляет существующую заметку пользователя
// @Tags         notes
// @Produce      json
// @Param        noteId  path  string  true  "ID заметки"
// @Success      204  "Заметка успешно удалена"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure      404  {object} dto.ErrorResponse "Заметка не найдена"
// @Failure      500  {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /notes/{noteId} [delete]
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	const op = "NoteHandler.DeleteNote"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	noteID, err := uuid.Parse(mux.Vars(r)["noteId"])
	if err != nil {
		logger.WithError(err).Warn("invalid note ID format")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid note ID format")
		return
	}

	err = h.uc.DeleteNote(r.Context(), noteID)
	if err != nil {
		logger.WithError(err).Error("failed to delete note")
		handler.HandleError(r.Context(), w, err, "Failed to delete note")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
