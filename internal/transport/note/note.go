package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lzimin05/course-todo/config"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/note"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/transport/utils/response"
)

type INoteUsecase interface {
	GetAllNotes(ctx context.Context) ([]dto.NoteDTO, error)
	CreateNote(ctx context.Context, req dto.CreateOrUpdateNote) (uuid.UUID, error)
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

func (h *NoteHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	const op = "NoteHandler.GetAllNotes"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	notes, err := h.uc.GetAllNotes(r.Context())
	if err != nil {
		logger.WithError(err).Error("failed to get notes")
		switch err {
		case errs.ErrEmptyNoteName:
			response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request data")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, notes)
}

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
		switch err {
		case errs.ErrEmptyNoteName:
			response.SendError(r.Context(), w, http.StatusBadRequest, "Note name cannot be empty")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to create note")
		}
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, map[string]interface{}{
		"id": noteID,
	})
}

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
		switch err {
		case errs.ErrEmptyNoteName:
			response.SendError(r.Context(), w, http.StatusBadRequest, "Note name cannot be empty")
		case errs.ErrNotFound:
			response.SendError(r.Context(), w, http.StatusNotFound, "Note not found")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to update note")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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
		switch err {
		case errs.ErrNotFound:
			response.SendError(r.Context(), w, http.StatusNotFound, "Note not found")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to delete note")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}