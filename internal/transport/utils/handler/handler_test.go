package handler

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/internal/models/errs"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func setupHandlerTest() context.Context {
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.PanicLevel) // Suppress log output during tests
	return logctx.WithLogger(context.Background(), logger)
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		defaultMsg     string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrNoAccess",
			err:            errs.ErrNoAccess,
			defaultMsg:     "Default message",
			expectedStatus: 403,
			expectedMsg:    "Access denied",
		},
		{
			name:           "ErrNotOwner",
			err:            errs.ErrNotOwner,
			defaultMsg:     "Default message",
			expectedStatus: 403,
			expectedMsg:    "Insufficient permissions",
		},
		{
			name:           "ErrOwnerCannotLeave",
			err:            errs.ErrOwnerCannotLeave,
			defaultMsg:     "Default message",
			expectedStatus: 403,
			expectedMsg:    "Project owner cannot leave project",
		},
		{
			name:           "ErrNotFound",
			err:            errs.ErrNotFound,
			defaultMsg:     "Default message",
			expectedStatus: 404,
			expectedMsg:    "Resource not found",
		},
		{
			name:           "ErrInvalidID",
			err:            errs.ErrInvalidID,
			defaultMsg:     "Default message",
			expectedStatus: 400,
			expectedMsg:    "Invalid ID format",
		},
		{
			name:           "Generic error",
			err:            errors.New("some generic error"),
			defaultMsg:     "Something went wrong",
			expectedStatus: 500,
			expectedMsg:    "Something went wrong",
		},
		{
			name:           "Nil error (should use default)",
			err:            nil,
			defaultMsg:     "Default error message",
			expectedStatus: 500,
			expectedMsg:    "Default error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupHandlerTest()
			w := httptest.NewRecorder()

			HandleError(ctx, w, tt.err, tt.defaultMsg)

			assert.Equal(t, tt.expectedStatus, w.Code)
			// Check that response contains the expected message (accounting for JSON escaping)
			body := w.Body.String()
			assert.Contains(t, body, "message")

			// For special characters test, check for JSON-escaped version
			if tt.name == "Default_message_with_special_characters" {
				assert.Contains(t, body, "\\u003c\\u003e\\u0026\\\"'")
			} else {
				assert.Contains(t, body, tt.expectedMsg)
			}
		})
	}
}

func TestHandleError_WrappedErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Wrapped ErrNoAccess",
			err:            errors.Join(errs.ErrNoAccess, errors.New("additional context")),
			expectedStatus: 403,
			expectedMsg:    "Access denied",
		},
		{
			name:           "Wrapped ErrNotOwner with context",
			err:            errors.Join(errors.New("context error"), errs.ErrNotOwner),
			expectedStatus: 403,
			expectedMsg:    "Insufficient permissions",
		},
		{
			name:           "Multiple wrapped ErrNotFound",
			err:            errors.Join(errs.ErrNotFound, errors.New("db error"), errors.New("network error")),
			expectedStatus: 404,
			expectedMsg:    "Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupHandlerTest()
			w := httptest.NewRecorder()

			HandleError(ctx, w, tt.err, "Default message")

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedMsg)
		})
	}
}

func TestHandleError_ErrorPriority(t *testing.T) {
	// Test that specific errors take priority over generic ones
	// when multiple errors are wrapped together
	ctx := setupHandlerTest()

	// Create an error that wraps both a specific error and a generic one
	specificErr := errs.ErrNoAccess
	genericErr := errors.New("generic database error")
	combinedErr := errors.Join(genericErr, specificErr)

	w := httptest.NewRecorder()
	HandleError(ctx, w, combinedErr, "Default message")

	// Should handle the specific error (ErrNoAccess) rather than falling back to default
	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "Access denied")
}

func TestHandleError_AllErrorTypes(t *testing.T) {
	// Comprehensive test to ensure all defined error types are handled
	ctx := setupHandlerTest()

	errorTests := map[error]int{
		errs.ErrNoAccess:         403,
		errs.ErrNotOwner:         403,
		errs.ErrOwnerCannotLeave: 403,
		errs.ErrNotFound:         404,
		errs.ErrInvalidID:        400,
	}

	for err, expectedStatus := range errorTests {
		t.Run(err.Error(), func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(ctx, w, err, "Default message")
			assert.Equal(t, expectedStatus, w.Code)
		})
	}
}

func TestHandleError_DefaultMessageVariations(t *testing.T) {
	tests := []struct {
		name       string
		defaultMsg string
	}{
		{
			name:       "Empty default message",
			defaultMsg: "",
		},
		{
			name:       "Long default message",
			defaultMsg: "This is a very long default error message that should still be handled properly",
		},
		{
			name:       "Default message with special characters",
			defaultMsg: "Error with special chars: <>&\"'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupHandlerTest()
			w := httptest.NewRecorder()

			// Use a generic error to trigger default message usage
			genericErr := errors.New("generic error")
			HandleError(ctx, w, genericErr, tt.defaultMsg)

			assert.Equal(t, 500, w.Code)
			body := w.Body.String()
			// For special characters test, check for JSON-escaped version
			if tt.name == "Default message with special characters" {
				assert.Contains(t, body, "\\u003c\\u003e\\u0026\\\"'")
			} else {
				assert.Contains(t, body, tt.defaultMsg)
			}
		})
	}
}
