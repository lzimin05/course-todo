package usecase

import (
	"context"

	"github.com/google/uuid"
	models "github.com/lzimin05/course-todo/internal/models/note"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/note"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/helpers"
)

type INoteRepository interface {
	GetAllNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uuid.UUID, name, description string) (uuid.UUID, error)
	UpdateNote(ctx context.Context, userID, noteID uuid.UUID, name, description string) error
	DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error
}

type NoteUsecase struct {
	repo INoteRepository
}

func NewNoteUsecase(repo INoteRepository) *NoteUsecase {
	return &NoteUsecase{
		repo: repo,
	}
}

func (u *NoteUsecase) GetAllNotes(ctx context.Context) ([]dto.NoteDTO, error) {
	const op = "NoteUsecase.GetAllNotes"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	notes, err := u.repo.GetAllNotes(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get notes from repository")
		return nil, err
	}

	var notesDTO []dto.NoteDTO
	for _, n := range notes {
		notesDTO = append(notesDTO, dto.NoteDTO{
			ID:          n.ID,
			Name:        n.Name,
			Description: n.Description,
		})
	}

	return notesDTO, nil
}

func (u *NoteUsecase) CreateNote(ctx context.Context, req dto.CreateOrUpdateNote) (uuid.UUID, error) {
	const op = "NoteUsecase.CreateNote"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return uuid.Nil, err
	}

	if req.Name == "" {
		logger.Warn("empty note name")
		return uuid.Nil, errs.ErrEmptyNoteName
	}

	noteID, err := u.repo.CreateNote(ctx, userID, req.Name, req.Description)
	if err != nil {
		logger.WithError(err).Error("failed to create note in repository")
		return uuid.Nil, err
	}

	return noteID, nil
}

func (u *NoteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, req dto.CreateOrUpdateNote) error {
	const op = "NoteUsecase.UpdateNote"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("noteID", noteID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return err
	}

	if req.Name == "" {
		logger.Warn("empty note name")
		return errs.ErrEmptyNoteName
	}

	err = u.repo.UpdateNote(ctx, userID, noteID, req.Name, req.Description)
	if err != nil {
		logger.WithError(err).Error("failed to update note in repository")
		return err
	}

	return nil
}

func (u *NoteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	const op = "NoteUsecase.DeleteNote"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("noteID", noteID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return err
	}

	err = u.repo.DeleteNote(ctx, userID, noteID)
	if err != nil {
		logger.WithError(err).Error("failed to delete note from repository")
		return err
	}

	return nil
}