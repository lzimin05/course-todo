package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	errs "github.com/lzimin05/course-todo/internal/models/errs"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		setupMocks  func()
		expectedErr error
		expectUser  bool
	}{
		{
			name:   "successful user retrieval by ID",
			userID: userID,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "username", "email", "password_hash"}).
					AddRow(userID, "testuser", "Test User", "test@example.com", []byte("hashedpassword"))

				mock.ExpectQuery(`SELECT`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedErr: nil,
			expectUser:  true,
		},
		{
			name:   "user not found",
			userID: userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: errs.ErrInvalidCredentials,
			expectUser:  false,
		},
		{
			name:   "database error",
			userID: userID,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs(userID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("database connection error"),
			expectUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := repo.GetUserByID(ctx, tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrInvalidCredentials {
					assert.Equal(t, errs.ErrInvalidCredentials, err)
				} else {
					assert.Contains(t, err.Error(), "database connection error")
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, userID, user.ID)
				assert.Equal(t, "testuser", user.Login)
				assert.Equal(t, "Test User", user.Username)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, []byte("hashedpassword"), user.PasswordHash)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()

	tests := []struct {
		name        string
		email       string
		setupMocks  func()
		expectedErr error
		expectUser  bool
	}{
		{
			name:  "successful user retrieval by email",
			email: "test@example.com",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "username", "email", "password_hash"}).
					AddRow(userID, "testuser", "Test User", "test@example.com", []byte("hashedpassword"))

				mock.ExpectQuery(`SELECT`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedErr: nil,
			expectUser:  true,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: errs.ErrInvalidCredentials,
			expectUser:  false,
		},
		{
			name:  "database error",
			email: "test@example.com",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs("test@example.com").
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("database connection error"),
			expectUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := repo.GetUserByEmail(ctx, tt.email)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrInvalidCredentials {
					assert.Equal(t, errs.ErrInvalidCredentials, err)
				} else {
					assert.Contains(t, err.Error(), "database connection error")
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, userID, user.ID)
				assert.Equal(t, "testuser", user.Login)
				assert.Equal(t, "Test User", user.Username)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, []byte("hashedpassword"), user.PasswordHash)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetUserByLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()

	tests := []struct {
		name        string
		login       string
		setupMocks  func()
		expectedErr error
		expectUser  bool
	}{
		{
			name:  "successful user retrieval by login",
			login: "testuser",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "username", "email", "password_hash"}).
					AddRow(userID, "testuser", "Test User", "test@example.com", []byte("hashedpassword"))

				mock.ExpectQuery(`SELECT`).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expectedErr: nil,
			expectUser:  true,
		},
		{
			name:  "user not found",
			login: "nonexistent",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: errs.ErrInvalidCredentials,
			expectUser:  false,
		},
		{
			name:  "database error",
			login: "testuser",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs("testuser").
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("database connection error"),
			expectUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := repo.GetUserByLogin(ctx, tt.login)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrInvalidCredentials {
					assert.Equal(t, errs.ErrInvalidCredentials, err)
				} else {
					assert.Contains(t, err.Error(), "database connection error")
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, userID, user.ID)
				assert.Equal(t, "testuser", user.Login)
				assert.Equal(t, "Test User", user.Username)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, []byte("hashedpassword"), user.PasswordHash)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNew_UserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
