package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	errs "github.com/lzimin05/course-todo/internal/models/errs"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func TestNoteRepository_GetNotesByProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	projectID := uuid.New()
	userID := uuid.New()
	noteID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name        string
		projectID   uuid.UUID
		userID      uuid.UUID
		setupMocks  func()
		expectedErr bool
		expectNotes int
	}{
		{
			name:      "successful notes retrieval",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "name", "description", "created_at"}).
					AddRow(noteID, projectID, userID, "Note 1", "Description 1", createdAt).
					AddRow(uuid.New(), projectID, userID, "Note 2", "Description 2", createdAt)

				mock.ExpectQuery(`SELECT n.id, n.project_id, n.user_id, n.name, n.description, n.created_at`).
					WithArgs(projectID, userID).
					WillReturnRows(rows)
			},
			expectedErr: false,
			expectNotes: 2,
		},
		{
			name:      "no notes found",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "name", "description", "created_at"})

				mock.ExpectQuery(`SELECT n.id, n.project_id, n.user_id, n.name, n.description, n.created_at`).
					WithArgs(projectID, userID).
					WillReturnRows(rows)
			},
			expectedErr: false,
			expectNotes: 0,
		},
		{
			name:      "database error",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT n.id, n.project_id, n.user_id, n.name, n.description, n.created_at`).
					WithArgs(projectID, userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
			expectNotes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			notes, err := repo.GetNotesByProject(ctx, tt.projectID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NoteRepository.GetNotesByProject")
				assert.Nil(t, notes)
			} else {
				assert.NoError(t, err)
				assert.Len(t, notes, tt.expectNotes)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNoteRepository_GetAllNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()
	projectID := uuid.New()
	noteID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name        string
		userID      uuid.UUID
		setupMocks  func()
		expectedErr bool
		expectNotes int
	}{
		{
			name:   "successful all notes retrieval",
			userID: userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "project_id", "user_id", "name", "description", "created_at"}).
					AddRow(noteID, projectID, userID, "User Note 1", "Description 1", createdAt)

				mock.ExpectQuery(`SELECT n.id, n.project_id, n.user_id, n.name, n.description, n.created_at`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedErr: false,
			expectNotes: 1,
		},
		{
			name:   "database error",
			userID: userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT n.id, n.project_id, n.user_id, n.name, n.description, n.created_at`).
					WithArgs(userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
			expectNotes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			notes, err := repo.GetAllNotes(ctx, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NoteRepository.GetAllNotes")
				assert.Nil(t, notes)
			} else {
				assert.NoError(t, err)
				assert.Len(t, notes, tt.expectNotes)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNoteRepository_CreateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	projectID := uuid.New()
	userID := uuid.New()
	noteID := uuid.New()

	tests := []struct {
		name        string
		projectID   uuid.UUID
		userID      uuid.UUID
		noteName    string
		description string
		setupMocks  func()
		expectedErr bool
	}{
		{
			name:        "successful note creation",
			projectID:   projectID,
			userID:      userID,
			noteName:    "Test Note",
			description: "Test Description",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(noteID)

				mock.ExpectQuery(`INSERT INTO todo.note`).
					WithArgs(sqlmock.AnyArg(), projectID, userID, "Test Note", "Test Description", sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:        "database constraint violation",
			projectID:   projectID,
			userID:      userID,
			noteName:    "Test Note",
			description: "Test Description",
			setupMocks: func() {
				pqErr := &pq.Error{Code: "23505"}
				mock.ExpectQuery(`INSERT INTO todo.note`).
					WithArgs(sqlmock.AnyArg(), projectID, userID, "Test Note", "Test Description", sqlmock.AnyArg()).
					WillReturnError(pqErr)
			},
			expectedErr: true,
		},
		{
			name:        "database error",
			projectID:   projectID,
			userID:      userID,
			noteName:    "Test Note",
			description: "Test Description",
			setupMocks: func() {
				mock.ExpectQuery(`INSERT INTO todo.note`).
					WithArgs(sqlmock.AnyArg(), projectID, userID, "Test Note", "Test Description", sqlmock.AnyArg()).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			id, err := repo.CreateNote(ctx, tt.projectID, tt.userID, tt.noteName, tt.description)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NoteRepository.CreateNote")
				assert.Equal(t, uuid.Nil, id)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNoteRepository_UpdateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()
	noteID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		noteID      uuid.UUID
		projectID   uuid.UUID
		noteName    string
		description string
		setupMocks  func()
		expectedErr error
	}{
		{
			name:        "successful note update",
			userID:      userID,
			noteID:      noteID,
			projectID:   projectID,
			noteName:    "Updated Note",
			description: "Updated Description",
			setupMocks: func() {
				mock.ExpectExec(`UPDATE todo.note`).
					WithArgs(noteID, projectID, userID, "Updated Note", "Updated Description").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name:        "note not found",
			userID:      userID,
			noteID:      noteID,
			projectID:   projectID,
			noteName:    "Updated Note",
			description: "Updated Description",
			setupMocks: func() {
				mock.ExpectExec(`UPDATE todo.note`).
					WithArgs(noteID, projectID, userID, "Updated Note", "Updated Description").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr: errs.ErrNotFound,
		},
		{
			name:        "database error",
			userID:      userID,
			noteID:      noteID,
			projectID:   projectID,
			noteName:    "Updated Note",
			description: "Updated Description",
			setupMocks: func() {
				mock.ExpectExec(`UPDATE todo.note`).
					WithArgs(noteID, projectID, userID, "Updated Note", "Updated Description").
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := repo.UpdateNote(ctx, tt.userID, tt.noteID, tt.projectID, tt.noteName, tt.description)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrNotFound {
					assert.Equal(t, errs.ErrNotFound, err)
				} else {
					assert.Contains(t, err.Error(), "NoteRepository.UpdateNote")
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNoteRepository_DeleteNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()
	noteID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		noteID      uuid.UUID
		setupMocks  func()
		expectedErr error
	}{
		{
			name:   "successful note deletion",
			userID: userID,
			noteID: noteID,
			setupMocks: func() {
				mock.ExpectExec(`DELETE FROM todo.note`).
					WithArgs(noteID, userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name:   "note not found",
			userID: userID,
			noteID: noteID,
			setupMocks: func() {
				mock.ExpectExec(`DELETE FROM todo.note`).
					WithArgs(noteID, userID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr: errs.ErrNotFound,
		},
		{
			name:   "database error",
			userID: userID,
			noteID: noteID,
			setupMocks: func() {
				mock.ExpectExec(`DELETE FROM todo.note`).
					WithArgs(noteID, userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := repo.DeleteNote(ctx, tt.userID, tt.noteID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrNotFound {
					assert.Equal(t, errs.ErrNotFound, err)
				} else {
					assert.Contains(t, err.Error(), "NoteRepository.DeleteNote")
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNewNoteRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
