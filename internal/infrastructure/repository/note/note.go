package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/note"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

const (
	getAllNotesByProjectQuery = `
		SELECT n.id, n.project_id, n.user_id, n.name, n.description, n.created_at
		FROM todo.note n
		JOIN todo.project_member pm ON n.project_id = pm.project_id
		WHERE n.project_id = $1 AND pm.user_id = $2`

	getAllNotesQuery = `
		SELECT n.id, n.project_id, n.user_id, n.name, n.description, n.created_at
		FROM todo.note n
		JOIN todo.project_member pm ON n.project_id = pm.project_id
		WHERE n.user_id = $1`

	createNoteQuery = `
		INSERT INTO todo.note (id, project_id, user_id, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	updateNoteQuery = `
		UPDATE todo.note 
		SET name = $4, description = $5 
		WHERE id = $1 AND project_id = $2 AND project_id IN (
			SELECT pm.project_id FROM todo.project_member pm WHERE pm.user_id = $3
		)`

	deleteNoteQuery = `
		DELETE FROM todo.note 
		WHERE id = $1 AND project_id IN (
			SELECT pm.project_id FROM todo.project_member pm WHERE pm.user_id = $2
		)`
)

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) GetNotesByProject(ctx context.Context, projectID, userID uuid.UUID) ([]models.Note, error) {
	const op = "NoteRepository.GetNotesByProject"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("projectID", projectID).
		WithField("userID", userID)

	rows, err := r.db.QueryContext(ctx, getAllNotesByProjectQuery, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get notes by project")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var n models.Note
		err := rows.Scan(&n.ID, &n.ProjectID, &n.UserID, &n.Name, &n.Description, &n.CreatedAt)
		if err != nil {
			logger.WithError(err).Error("failed to scan note")
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		notes = append(notes, n)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("rows iteration error")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return notes, nil
}

func (r *NoteRepository) GetAllNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	const op = "NoteRepository.GetAllNotes"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("userID", userID)

	rows, err := r.db.QueryContext(ctx, getAllNotesQuery, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get notes")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var n models.Note
		err := rows.Scan(&n.ID, &n.ProjectID, &n.UserID, &n.Name, &n.Description, &n.CreatedAt)
		if err != nil {
			logger.WithError(err).Error("failed to scan note")
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		notes = append(notes, n)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("rows iteration error")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return notes, nil
}

func (r *NoteRepository) CreateNote(ctx context.Context, projectID, userID uuid.UUID, name, description string) (uuid.UUID, error) {
	const op = "NoteRepository.CreateNote"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("projectID", projectID).
		WithField("userID", userID)

	newNote := models.Note{
		ID:          uuid.New(),
		ProjectID:   projectID,
		UserID:      userID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, createNoteQuery,
		newNote.ID, newNote.ProjectID, newNote.UserID, newNote.Name, newNote.Description, newNote.CreatedAt).
		Scan(&id)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			logger.WithError(pqErr).Error("database error")
			return uuid.Nil, fmt.Errorf("%s: %w", op, pqErr)
		}
		logger.WithError(err).Error("failed to create note")
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *NoteRepository) UpdateNote(ctx context.Context, userID, noteID, projectID uuid.UUID, name, description string) error {
	const op = "NoteRepository.UpdateNote"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("userID", userID).
		WithField("noteID", noteID).
		WithField("projectID", projectID)

	result, err := r.db.ExecContext(ctx, updateNoteQuery,
		noteID, projectID, userID, name, description)

	if err != nil {
		logger.WithError(err).Error("failed to update note")
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("failed to get rows affected")
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		logger.Warn("note not found or not owned by user")
		return errs.ErrNotFound
	}

	return nil
}

func (r *NoteRepository) DeleteNote(ctx context.Context, userID, noteID uuid.UUID) error {
	const op = "NoteRepository.DeleteNote"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("userID", userID).
		WithField("noteID", noteID)

	result, err := r.db.ExecContext(ctx, deleteNoteQuery, noteID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to delete note")
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("failed to get rows affected")
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		logger.Warn("note not found or not owned by user")
		return errs.ErrNotFound
	}

	return nil
}
