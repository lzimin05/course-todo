package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/project"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func TestProjectRepository_CreateProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	projectID := uuid.New()
	ownerID := uuid.New()
	memberID := uuid.New()
	createdAt := time.Now()
	joinedAt := time.Now()

	project := &models.Project{
		Name:        "Test Project",
		Description: "Test Description",
		OwnerID:     ownerID,
	}

	tests := []struct {
		name        string
		project     *models.Project
		setupMocks  func()
		expectedErr bool
	}{
		{
			name:    "successful project creation",
			project: project,
			setupMocks: func() {
				mock.ExpectBegin()

				// Mock project creation
				rows := sqlmock.NewRows([]string{"id", "created_at"}).
					AddRow(projectID, createdAt)
				mock.ExpectQuery(`INSERT INTO todo.project`).
					WithArgs("Test Project", "Test Description", ownerID).
					WillReturnRows(rows)

				// Mock owner as member addition
				memberRows := sqlmock.NewRows([]string{"id", "joined_at"}).
					AddRow(memberID, joinedAt)
				mock.ExpectQuery(`INSERT INTO todo.project_member`).
					WithArgs(projectID, ownerID, models.RoleOwner).
					WillReturnRows(memberRows)

				mock.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name:    "transaction begin error",
			project: project,
			setupMocks: func() {
				mock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))
			},
			expectedErr: true,
		},
		{
			name:    "project creation error",
			project: project,
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO todo.project`).
					WithArgs("Test Project", "Test Description", ownerID).
					WillReturnError(errors.New("project creation error"))
				mock.ExpectRollback()
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := repo.CreateProject(ctx, tt.project)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.project.ID)
				assert.NotZero(t, tt.project.CreatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProjectRepository_GetProjectByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	projectID := uuid.New()
	ownerID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name          string
		projectID     uuid.UUID
		setupMocks    func()
		expectedErr   error
		expectProject bool
	}{
		{
			name:      "successful project retrieval",
			projectID: projectID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "created_at"}).
					AddRow(projectID, "Test Project", "Test Description", ownerID, createdAt)

				mock.ExpectQuery(`SELECT p.id, p.name, p.description, p.owner_id, p.created_at`).
					WithArgs(projectID).
					WillReturnRows(rows)
			},
			expectedErr:   nil,
			expectProject: true,
		},
		{
			name:      "project not found",
			projectID: projectID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT p.id, p.name, p.description, p.owner_id, p.created_at`).
					WithArgs(projectID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr:   errs.ErrNotFound,
			expectProject: false,
		},
		{
			name:      "database error",
			projectID: projectID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT p.id, p.name, p.description, p.owner_id, p.created_at`).
					WithArgs(projectID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr:   errors.New("database connection error"),
			expectProject: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			project, err := repo.GetProjectByID(ctx, tt.projectID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrNotFound {
					assert.Equal(t, errs.ErrNotFound, err)
				}
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, projectID, project.ID)
				assert.Equal(t, "Test Project", project.Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProjectRepository_GetUserProjects(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()
	projectID := uuid.New()
	ownerID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name           string
		userID         uuid.UUID
		setupMocks     func()
		expectedErr    bool
		expectProjects int
	}{
		{
			name:   "successful user projects retrieval",
			userID: userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "created_at"}).
					AddRow(projectID, "User Project 1", "Description 1", ownerID, createdAt).
					AddRow(uuid.New(), "User Project 2", "Description 2", ownerID, createdAt)

				mock.ExpectQuery(`SELECT p.id, p.name, p.description, p.owner_id, p.created_at`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedErr:    false,
			expectProjects: 2,
		},
		{
			name:   "no projects found",
			userID: userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "created_at"})

				mock.ExpectQuery(`SELECT p.id, p.name, p.description, p.owner_id, p.created_at`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedErr:    false,
			expectProjects: 0,
		},
		{
			name:   "database error",
			userID: userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT p.id, p.name, p.description, p.owner_id, p.created_at`).
					WithArgs(userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr:    true,
			expectProjects: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			projects, err := repo.GetUserProjects(ctx, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, projects)
			} else {
				assert.NoError(t, err)
				assert.Len(t, projects, tt.expectProjects)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProjectRepository_CheckProjectAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	projectID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name         string
		projectID    uuid.UUID
		userID       uuid.UUID
		setupMocks   func()
		expectedErr  bool
		expectAccess bool
	}{
		{
			name:      "user has access",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs(projectID, userID).
					WillReturnRows(rows)
			},
			expectedErr:  false,
			expectAccess: true,
		},
		{
			name:      "user has no access",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs(projectID, userID).
					WillReturnRows(rows)
			},
			expectedErr:  false,
			expectAccess: false,
		},
		{
			name:      "database error",
			projectID: projectID,
			userID:    userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs(projectID, userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr:  true,
			expectAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			hasAccess, err := repo.CheckProjectAccess(ctx, tt.projectID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.False(t, hasAccess)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectAccess, hasAccess)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProjectRepository_DeleteProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	projectID := uuid.New()
	ownerID := uuid.New()

	tests := []struct {
		name        string
		projectID   uuid.UUID
		ownerID     uuid.UUID
		setupMocks  func()
		expectedErr error
	}{
		{
			name:      "successful project deletion",
			projectID: projectID,
			ownerID:   ownerID,
			setupMocks: func() {
				mock.ExpectExec(`DELETE FROM todo.project`).
					WithArgs(projectID, ownerID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name:      "project not found",
			projectID: projectID,
			ownerID:   ownerID,
			setupMocks: func() {
				mock.ExpectExec(`DELETE FROM todo.project`).
					WithArgs(projectID, ownerID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr: errs.ErrNotFound,
		},
		{
			name:      "database error",
			projectID: projectID,
			ownerID:   ownerID,
			setupMocks: func() {
				mock.ExpectExec(`DELETE FROM todo.project`).
					WithArgs(projectID, ownerID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := repo.DeleteProject(ctx, tt.projectID, tt.ownerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrNotFound {
					assert.Equal(t, errs.ErrNotFound, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNew_ProjectRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
