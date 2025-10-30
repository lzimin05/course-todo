package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	models "github.com/lzimin05/course-todo/internal/models/note"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lib/pq"
)

const (
	getAllNotesQuery = `
		SELECT id, name, description 
		FROM todo.note 
		WHERE user_id = $1`

	createNoteQuery = `
		INSERT INTO todo.note (id, user_id, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	updateNoteQuery = `
		UPDATE todo.note 
		SET name = $3, description = $4 
		WHERE id = $1 AND user_id = $2`

	deleteNoteQuery = `
		DELETE FROM todo.note 
		WHERE id = $1 AND user_id = $2`
)

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
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
		err := rows.Scan(&n.ID, &n.Name, &n.Description)
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

func (r *NoteRepository) CreateNote(ctx context.Context, userID uuid.UUID, name, description string) (uuid.UUID, error) {
	const op = "NoteRepository.CreateNote"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("userID", userID)

	newNote := models.Note{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, createNoteQuery,
		newNote.ID, newNote.UserID, newNote.Name, newNote.Description, newNote.CreatedAt).
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

func (r *NoteRepository) UpdateNote(ctx context.Context, userID, noteID uuid.UUID, name, description string) error {
	const op = "NoteRepository.UpdateNote"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("userID", userID).
		WithField("noteID", noteID)

	result, err := r.db.ExecContext(ctx, updateNoteQuery,
		noteID, userID, name, description)

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