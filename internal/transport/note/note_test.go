package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/models/domains"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/note"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/mocks"
)

func TestNoteHandler_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockINoteUsecase(ctrl)
	config := &config.Config{}
	handler := NewNoteHandler(mockUsecase, config)

	notes := []*dto.NoteDTO{
		{
			ID:          uuid.New(),
			ProjectID:   uuid.New(),
			UserID:      uuid.New(),
			Name:        "Test Note 1",
			Description: "Description 1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			ProjectID:   uuid.New(),
			UserID:      uuid.New(),
			Name:        "Test Note 2",
			Description: "Description 2",
			CreatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name           string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful retrieval",
			setupMocks: func() {
				mockUsecase.EXPECT().GetAllNotes(gomock.Any()).Return(notes, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "usecase error",
			setupMocks: func() {
				mockUsecase.EXPECT().GetAllNotes(gomock.Any()).Return(nil, errs.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/notes/all", nil)
			ctx := req.Context()
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			ctx = context.WithValue(ctx, domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetAllNotes(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var result []*dto.NoteDTO
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Len(t, result, len(notes))
			}
		})
	}
}

func TestNoteHandler_GetNotesByProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockINoteUsecase(ctrl)
	config := &config.Config{}
	handler := NewNoteHandler(mockUsecase, config)

	projectID := uuid.New()
	notes := []*dto.NoteDTO{
		{
			ID:          uuid.New(),
			ProjectID:   projectID,
			UserID:      uuid.New(),
			Name:        "Project Note",
			Description: "Description",
			CreatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name           string
		projectID      string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:      "successful retrieval",
			projectID: projectID.String(),
			setupMocks: func() {
				mockUsecase.EXPECT().GetNotesByProject(gomock.Any(), projectID).Return(notes, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:      "invalid project ID",
			projectID: "invalid-uuid",
			setupMocks: func() {
				// No mock expectations as validation happens before usecase call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:      "no access to project",
			projectID: projectID.String(),
			setupMocks: func() {
				mockUsecase.EXPECT().GetNotesByProject(gomock.Any(), projectID).Return(nil, errs.ErrNoAccess)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/notes/project/%s", tt.projectID), nil)
			ctx := req.Context()
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			ctx = context.WithValue(ctx, domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			// Setup mux vars
			req = mux.SetURLVars(req, map[string]string{"projectId": tt.projectID})

			rr := httptest.NewRecorder()
			handler.GetNotesByProject(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var result []*dto.NoteDTO
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Len(t, result, len(notes))
			}
		})
	}
}

func TestNoteHandler_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockINoteUsecase(ctrl)
	config := &config.Config{}
	handler := NewNoteHandler(mockUsecase, config)

	projectID := uuid.New()
	noteID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful creation",
			requestBody: dto.CreateOrUpdateNote{
				ProjectID:   projectID,
				Name:        "New Note",
				Description: "Note Description",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(&dto.CreateNoteDTO{ID: noteID}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:        "invalid JSON",
			requestBody: "invalid json",
			setupMocks: func() {
				// No mock expectations as validation happens before usecase call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "validation error - missing project ID",
			requestBody: dto.CreateOrUpdateNote{
				Name:        "New Note",
				Description: "Note Description",
			},
			setupMocks: func() {
				// Usecase will be called but will return validation error
				mockUsecase.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil, errs.ErrNoAccess)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  true,
		},
		{
			name: "usecase error",
			requestBody: dto.CreateOrUpdateNote{
				ProjectID:   projectID,
				Name:        "New Note",
				Description: "Note Description",
			},
			setupMocks: func() {
				expectedReq := dto.CreateOrUpdateNote{
					ProjectID:   projectID,
					Name:        "New Note",
					Description: "Note Description",
				}
				mockUsecase.EXPECT().CreateNote(gomock.Any(), expectedReq).Return(nil, errs.ErrNoAccess)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				var err error
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(reqBody))
			ctx := req.Context()
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			ctx = context.WithValue(ctx, domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.CreateNote(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var result dto.CreateNoteDTO
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, noteID, result.ID)
			}
		})
	}
}

func TestNoteHandler_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockINoteUsecase(ctrl)
	config := &config.Config{}
	handler := NewNoteHandler(mockUsecase, config)

	noteID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name           string
		noteID         string
		requestBody    interface{}
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "successful update",
			noteID: noteID.String(),
			requestBody: dto.CreateOrUpdateNote{
				ProjectID:   projectID,
				Name:        "Updated Note",
				Description: "Updated Description",
			},
			setupMocks: func() {
				expectedReq := dto.CreateOrUpdateNote{
					ProjectID:   projectID,
					Name:        "Updated Note",
					Description: "Updated Description",
				}
				mockUsecase.EXPECT().UpdateNote(gomock.Any(), noteID, expectedReq).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedError:  false,
		},
		{
			name:   "invalid note ID",
			noteID: "invalid-uuid",
			requestBody: dto.CreateOrUpdateNote{
				ProjectID:   projectID,
				Name:        "Updated Note",
				Description: "Updated Description",
			},
			setupMocks: func() {
				// No mock expectations as validation happens before usecase call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "invalid JSON",
			noteID:      noteID.String(),
			requestBody: "invalid json",
			setupMocks: func() {
				// No mock expectations as validation happens before usecase call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				var err error
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/notes/%s", tt.noteID), bytes.NewReader(reqBody))
			ctx := req.Context()
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			ctx = context.WithValue(ctx, domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			// Setup mux vars
			req = mux.SetURLVars(req, map[string]string{"noteId": tt.noteID})

			rr := httptest.NewRecorder()
			handler.UpdateNote(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestNoteHandler_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockINoteUsecase(ctrl)
	config := &config.Config{}
	handler := NewNoteHandler(mockUsecase, config)

	noteID := uuid.New()

	tests := []struct {
		name           string
		noteID         string
		setupMocks     func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "successful deletion",
			noteID: noteID.String(),
			setupMocks: func() {
				mockUsecase.EXPECT().DeleteNote(gomock.Any(), noteID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedError:  false,
		},
		{
			name:   "invalid note ID",
			noteID: "invalid-uuid",
			setupMocks: func() {
				// No mock expectations as validation happens before usecase call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "usecase error",
			noteID: noteID.String(),
			setupMocks: func() {
				mockUsecase.EXPECT().DeleteNote(gomock.Any(), noteID).Return(errs.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%s", tt.noteID), nil)
			ctx := req.Context()
			ctx = logctx.WithLogger(ctx, logctx.NewLogger())
			ctx = context.WithValue(ctx, domains.UserIDKey{}, uuid.New().String())
			req = req.WithContext(ctx)

			// Setup mux vars
			req = mux.SetURLVars(req, map[string]string{"noteId": tt.noteID})

			rr := httptest.NewRecorder()
			handler.DeleteNote(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestNewNoteHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockINoteUsecase(ctrl)
	config := &config.Config{}

	handler := NewNoteHandler(mockUsecase, config)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecase, handler.uc)
	assert.Equal(t, config, handler.config)
}
