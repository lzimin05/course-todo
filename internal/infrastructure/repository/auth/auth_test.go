package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	errs "github.com/lzimin05/course-todo/internal/models/errs"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func TestAuthRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	tests := []struct {
		name         string
		login        string
		username     string
		email        string
		passwordHash []byte
		setupMocks   func()
		expectedErr  error
	}{
		{
			name:         "successful user creation",
			login:        "testuser",
			username:     "Test User",
			email:        "test@example.com",
			passwordHash: []byte("hashedpassword"),
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "email", "username"}).
					AddRow(uuid.New(), "testuser", "test@example.com", "Test User")

				mock.ExpectQuery(`INSERT INTO todo."user"`).
					WithArgs(sqlmock.AnyArg(), "testuser", "Test User", "test@example.com", []byte("hashedpassword")).
					WillReturnRows(rows)
			},
			expectedErr: nil,
		},
		{
			name:         "duplicate key constraint violation",
			login:        "existinguser",
			username:     "Existing User",
			email:        "existing@example.com",
			passwordHash: []byte("hashedpassword"),
			setupMocks: func() {
				pqErr := &pq.Error{Code: "23505"}
				mock.ExpectQuery(`INSERT INTO todo."user"`).
					WithArgs(sqlmock.AnyArg(), "existinguser", "Existing User", "existing@example.com", []byte("hashedpassword")).
					WillReturnError(pqErr)
			},
			expectedErr: errs.ErrIsDuplicateKey,
		},
		{
			name:         "database error",
			login:        "testuser",
			username:     "Test User",
			email:        "test@example.com",
			passwordHash: []byte("hashedpassword"),
			setupMocks: func() {
				mock.ExpectQuery(`INSERT INTO todo."user"`).
					WithArgs(sqlmock.AnyArg(), "testuser", "Test User", "test@example.com", []byte("hashedpassword")).
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("AuthRepository.CreateUser: database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := repo.CreateUser(ctx, tt.login, tt.username, tt.email, tt.passwordHash)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == errs.ErrIsDuplicateKey {
					assert.Equal(t, errs.ErrIsDuplicateKey, err)
				} else {
					assert.Contains(t, err.Error(), "AuthRepository.CreateUser")
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.login, user.Login)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.passwordHash, user.PasswordHash)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_GetUserByEmail(t *testing.T) {
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
			name:  "successful user retrieval",
			email: "test@example.com",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "username", "email", "password_hash"}).
					AddRow(userID, "testuser", "Test User", "test@example.com", []byte("hashedpassword"))

				mock.ExpectQuery(`SELECT id, login, username, email, password_hash FROM todo."user" WHERE email = \$1`).
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
				mock.ExpectQuery(`SELECT id, login, username, email, password_hash FROM todo."user" WHERE email = \$1`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: nil,
			expectUser:  false,
		},
		{
			name:  "database error",
			email: "test@example.com",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT id, login, username, email, password_hash FROM todo."user" WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("AuthRepository.GetUserByEmail: database connection error"),
			expectUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := repo.GetUserByEmail(ctx, tt.email)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "AuthRepository.GetUserByEmail")
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tt.expectUser {
					assert.NotNil(t, user)
					assert.Equal(t, userID, user.ID)
					assert.Equal(t, "testuser", user.Login)
					assert.Equal(t, "Test User", user.Username)
					assert.Equal(t, tt.email, user.Email)
					assert.Equal(t, []byte("hashedpassword"), user.PasswordHash)
				} else {
					assert.Nil(t, user)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthRepository_GetUserByEmailOrLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := logctx.WithLogger(context.Background(), logctx.NewLogger())

	userID := uuid.New()

	tests := []struct {
		name        string
		identifier  string
		setupMocks  func()
		expectedErr error
		expectUser  bool
	}{
		{
			name:       "successful user retrieval by email",
			identifier: "test@example.com",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "username", "email", "password_hash"}).
					AddRow(userID, "testuser", "Test User", "test@example.com", []byte("hashedpassword"))

				mock.ExpectQuery(`SELECT id, login, username, email, password_hash FROM todo."user" WHERE email = \$1 or login = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedErr: nil,
			expectUser:  true,
		},
		{
			name:       "successful user retrieval by login",
			identifier: "testuser",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "login", "username", "email", "password_hash"}).
					AddRow(userID, "testuser", "Test User", "test@example.com", []byte("hashedpassword"))

				mock.ExpectQuery(`SELECT id, login, username, email, password_hash FROM todo."user" WHERE email = \$1 or login = \$1`).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expectedErr: nil,
			expectUser:  true,
		},
		{
			name:       "user not found",
			identifier: "nonexistent",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT id, login, username, email, password_hash FROM todo."user" WHERE email = \$1 or login = \$1`).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: nil,
			expectUser:  false,
		},
		{
			name:       "database error",
			identifier: "testuser",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT id, login, username, email, password_hash FROM todo."user" WHERE email = \$1 or login = \$1`).
					WithArgs("testuser").
					WillReturnError(errors.New("database connection error"))
			},
			expectedErr: errors.New("AuthRepository.GetUserByEmail: database connection error"),
			expectUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			user, err := repo.GetUserByEmailOrLogin(ctx, tt.identifier)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "AuthRepository.GetUserByEmail")
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tt.expectUser {
					assert.NotNil(t, user)
					assert.Equal(t, userID, user.ID)
					assert.Equal(t, "testuser", user.Login)
					assert.Equal(t, "Test User", user.Username)
					assert.Equal(t, "test@example.com", user.Email)
					assert.Equal(t, []byte("hashedpassword"), user.PasswordHash)
				} else {
					assert.Nil(t, user)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNew_AuthRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
