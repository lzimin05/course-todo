package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/note"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/note"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/helpers"
)

type INoteRepository interface {
	GetAllNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	GetNotesByProject(ctx context.Context, projectID, userID uuid.UUID) ([]models.Note, error)
	CreateNote(ctx context.Context, projectID, userID uuid.UUID, name, description string) (uuid.UUID, error)
	UpdateNote(ctx context.Context, userID, noteID, projectID uuid.UUID, name, description string) error
	DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error
}

type ProjectRepository interface {
	CheckProjectAccess(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
}

type NoteUsecase struct {
	repo        INoteRepository
	projectRepo ProjectRepository
}

func NewNoteUsecase(repo INoteRepository, projectRepo ProjectRepository) *NoteUsecase {
	return &NoteUsecase{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

func (u *NoteUsecase) GetAllNotes(ctx context.Context) ([]*dto.NoteDTO, error) {
	const op = "NoteUsecase.GetAllNotes"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	notesmodel, err := u.repo.GetAllNotes(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get notes from repository")
		return nil, err
	}

	notesDTO := make([]*dto.NoteDTO, len(notesmodel))
	for i, notemodel := range notesmodel {
		notesDTO[i] = &dto.NoteDTO{
			ID:          notemodel.ID,
			ProjectID:   notemodel.ProjectID,
			UserID:      notemodel.UserID,
			Name:        notemodel.Name,
			Description: notemodel.Description,
			CreatedAt:   notemodel.CreatedAt.Truncate(time.Second),
		}
	}

	return notesDTO, nil
}

func (u *NoteUsecase) GetNotesByProject(ctx context.Context, projectID uuid.UUID) ([]*dto.NoteDTO, error) {
	const op = "NoteUsecase.GetNotesByProject"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	// Проверяем права доступа к проекту
	hasAccess, err := u.projectRepo.CheckProjectAccess(ctx, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to check project access")
		return nil, err
	}

	if !hasAccess {
		logger.Warn("user doesn't have access to project")
		return nil, errs.ErrNoAccess
	}

	notesmodel, err := u.repo.GetNotesByProject(ctx, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get notes by project from repository")
		return nil, err
	}

	notesDTO := make([]*dto.NoteDTO, len(notesmodel))
	for i, notemodel := range notesmodel {
		notesDTO[i] = &dto.NoteDTO{
			ID:          notemodel.ID,
			ProjectID:   notemodel.ProjectID,
			UserID:      notemodel.UserID,
			Name:        notemodel.Name,
			Description: notemodel.Description,
			CreatedAt:   notemodel.CreatedAt.Truncate(time.Second),
		}
	}

	return notesDTO, nil
}

func (u *NoteUsecase) CreateNote(ctx context.Context, req dto.CreateOrUpdateNote) (*dto.CreateNoteDTO, error) {
	const op = "NoteUsecase.CreateNote"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	if req.Name == "" {
		logger.Warn("empty note name")
		return nil, errs.ErrEmptyNoteName
	}

	// Проверяем права доступа к проекту
	hasAccess, err := u.projectRepo.CheckProjectAccess(ctx, req.ProjectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to check project access")
		return nil, err
	}

	if !hasAccess {
		logger.Warn("user doesn't have access to project")
		return nil, errs.ErrNoAccess
	}

	noteID, err := u.repo.CreateNote(ctx, req.ProjectID, userID, req.Name, req.Description)
	if err != nil {
		logger.WithError(err).Error("failed to create note in repository")
		return nil, err
	}

	return &dto.CreateNoteDTO{
		ID: noteID,
	}, nil
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

	err = u.repo.UpdateNote(ctx, userID, noteID, req.ProjectID, req.Name, req.Description)
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
