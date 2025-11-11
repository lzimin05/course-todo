package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/internal/models/domains"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/note"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/note"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func setupNoteTest() (context.Context, uuid.UUID) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
	ctx = logctx.WithLogger(ctx, logctx.NewLogger())
	return ctx, userID
}

func TestNoteUsecase_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	noteRepo := mocks.NewMockINoteRepository(ctrl)
	projectRepo := mocks.NewMockNoteProjectRepository(ctrl)

	notes := []models.Note{
		{
			ID:          uuid.New(),
			Name:        "Note 1",
			Description: "Description 1",
			ProjectID:   uuid.New(),
			UserID:      uuid.New(),
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Note 2",
			Description: "Description 2",
			ProjectID:   uuid.New(),
			UserID:      uuid.New(),
			CreatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name        string
		setupMocks  func(uuid.UUID)
		expectedErr error
	}{
		{
			name: "successful retrieval",
			setupMocks: func(userID uuid.UUID) {
				noteRepo.EXPECT().GetAllNotes(gomock.Any(), userID).Return(notes, nil)
			},
			expectedErr: nil,
		},
		{
			name: "repository error",
			setupMocks: func(userID uuid.UUID) {
				noteRepo.EXPECT().GetAllNotes(gomock.Any(), userID).Return(nil, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, userID := setupNoteTest()
			tt.setupMocks(userID)

			uc := NewNoteUsecase(noteRepo, projectRepo)
			result, err := uc.GetAllNotes(ctx)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, len(notes))
			}
		})
	}
}

func TestNoteUsecase_GetNotesByProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	noteRepo := mocks.NewMockINoteRepository(ctrl)
	projectRepo := mocks.NewMockNoteProjectRepository(ctrl)

	projectID := uuid.New()
	userID := uuid.New()
	notes := []models.Note{
		{
			ID:          uuid.New(),
			Name:        "Project Note",
			Description: "Description",
			ProjectID:   projectID,
			UserID:      userID,
			CreatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name        string
		setupMocks  func()
		expectedErr error
	}{
		{
			name: "successful retrieval with access",
			setupMocks: func() {
				projectRepo.EXPECT().CheckProjectAccess(gomock.Any(), projectID, userID).Return(true, nil)
				noteRepo.EXPECT().GetNotesByProject(gomock.Any(), projectID, userID).Return(notes, nil)
			},
			expectedErr: nil,
		},
		{
			name: "no project access",
			setupMocks: func() {
				projectRepo.EXPECT().CheckProjectAccess(gomock.Any(), projectID, userID).Return(false, nil)
			},
			expectedErr: errs.ErrNoAccess,
		},
		{
			name: "project access check error",
			setupMocks: func() {
				projectRepo.EXPECT().CheckProjectAccess(gomock.Any(), projectID, userID).Return(false, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			tt.setupMocks()

			uc := NewNoteUsecase(noteRepo, projectRepo)
			result, err := uc.GetNotesByProject(ctx, projectID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestNoteUsecase_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	noteRepo := mocks.NewMockINoteRepository(ctrl)
	projectRepo := mocks.NewMockNoteProjectRepository(ctrl)

	projectID := uuid.New()
	userID := uuid.New()
	noteID := uuid.New()

	tests := []struct {
		name        string
		req         dto.CreateOrUpdateNote
		setupMocks  func()
		expectedErr error
	}{
		{
			name: "successful creation with access",
			req: dto.CreateOrUpdateNote{
				Name:        "New Note",
				Description: "Note Description",
				ProjectID:   projectID,
			},
			setupMocks: func() {
				projectRepo.EXPECT().CheckProjectAccess(gomock.Any(), projectID, userID).Return(true, nil)
				noteRepo.EXPECT().CreateNote(gomock.Any(), projectID, userID, "New Note", "Note Description").Return(noteID, nil)
			},
			expectedErr: nil,
		},
		{
			name: "no project access",
			req: dto.CreateOrUpdateNote{
				Name:        "New Note",
				Description: "Note Description",
				ProjectID:   projectID,
			},
			setupMocks: func() {
				projectRepo.EXPECT().CheckProjectAccess(gomock.Any(), projectID, userID).Return(false, nil)
			},
			expectedErr: errs.ErrNoAccess,
		},
		{
			name: "empty note name",
			req: dto.CreateOrUpdateNote{
				Name:        "",
				Description: "Note Description",
				ProjectID:   projectID,
			},
			setupMocks: func() {
				// No mocks needed as validation happens before repository calls
			},
			expectedErr: errs.ErrEmptyNoteName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			tt.setupMocks()

			uc := NewNoteUsecase(noteRepo, projectRepo)
			result, err := uc.CreateNote(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, noteID, result.ID)
			}
		})
	}
}

func TestNoteUsecase_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	noteRepo := mocks.NewMockINoteRepository(ctrl)
	projectRepo := mocks.NewMockNoteProjectRepository(ctrl)

	projectID := uuid.New()
	noteID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		req         dto.CreateOrUpdateNote
		setupMocks  func()
		expectedErr error
	}{
		{
			name: "successful update",
			req: dto.CreateOrUpdateNote{
				Name:        "Updated Note",
				Description: "Updated Description",
				ProjectID:   projectID,
			},
			setupMocks: func() {
				noteRepo.EXPECT().UpdateNote(gomock.Any(), userID, noteID, projectID, "Updated Note", "Updated Description").Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "empty note name",
			req: dto.CreateOrUpdateNote{
				Name:        "",
				Description: "Updated Description",
				ProjectID:   projectID,
			},
			setupMocks: func() {
				// No mocks needed as validation happens before repository calls
			},
			expectedErr: errs.ErrEmptyNoteName,
		},
		{
			name: "repository error",
			req: dto.CreateOrUpdateNote{
				Name:        "Updated Note",
				Description: "Updated Description",
				ProjectID:   projectID,
			},
			setupMocks: func() {
				noteRepo.EXPECT().UpdateNote(gomock.Any(), userID, noteID, projectID, "Updated Note", "Updated Description").Return(errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			tt.setupMocks()

			uc := NewNoteUsecase(noteRepo, projectRepo)
			err := uc.UpdateNote(ctx, noteID, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNoteUsecase_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	noteRepo := mocks.NewMockINoteRepository(ctrl)
	projectRepo := mocks.NewMockNoteProjectRepository(ctrl)

	noteID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		setupMocks  func()
		expectedErr error
	}{
		{
			name: "successful deletion",
			setupMocks: func() {
				noteRepo.EXPECT().DeleteNote(gomock.Any(), userID, noteID).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "repository error",
			setupMocks: func() {
				noteRepo.EXPECT().DeleteNote(gomock.Any(), userID, noteID).Return(errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), domains.UserIDKey{}, userID.String())
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			tt.setupMocks()

			uc := NewNoteUsecase(noteRepo, projectRepo)
			err := uc.DeleteNote(ctx, noteID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew_NoteUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	noteRepo := mocks.NewMockINoteRepository(ctrl)
	projectRepo := mocks.NewMockNoteProjectRepository(ctrl)

	uc := NewNoteUsecase(noteRepo, projectRepo)

	assert.NotNil(t, uc)
	assert.Equal(t, noteRepo, uc.repo)
	assert.Equal(t, projectRepo, uc.projectRepo)
}
