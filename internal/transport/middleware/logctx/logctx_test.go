package logctx

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/internal/models/domains"
)

func TestWithLogger(t *testing.T) {
	logger := logrus.NewEntry(logrus.New()).WithField("test", "value")
	ctx := context.Background()

	newCtx := WithLogger(ctx, logger)

	// Check that the logger is properly stored in context
	storedLogger := newCtx.Value(domains.LoggerKey{})
	assert.NotNil(t, storedLogger)
	assert.Equal(t, logger, storedLogger)
}

func TestGetLogger_WithValidLogger(t *testing.T) {
	originalLogger := logrus.NewEntry(logrus.New()).WithFields(logrus.Fields{
		"request_id": "test-123",
		"operation":  "test",
	})

	ctx := WithLogger(context.Background(), originalLogger)

	retrievedLogger := GetLogger(ctx)

	assert.NotNil(t, retrievedLogger)
	assert.Equal(t, originalLogger, retrievedLogger)

	// Check that the fields are preserved
	assert.Equal(t, "test-123", retrievedLogger.Data["request_id"])
	assert.Equal(t, "test", retrievedLogger.Data["operation"])
}

func TestGetLogger_WithoutLogger(t *testing.T) {
	ctx := context.Background() // No logger in context

	retrievedLogger := GetLogger(ctx)

	assert.NotNil(t, retrievedLogger)
	assert.IsType(t, &logrus.Entry{}, retrievedLogger)

	// Should return a new default logger
	assert.Empty(t, retrievedLogger.Data)
}

func TestGetLogger_WithInvalidType(t *testing.T) {
	ctx := context.WithValue(context.Background(), domains.LoggerKey{}, "not a logger")

	retrievedLogger := GetLogger(ctx)

	assert.NotNil(t, retrievedLogger)
	assert.IsType(t, &logrus.Entry{}, retrievedLogger)

	// Should return a new default logger when type assertion fails
	assert.Empty(t, retrievedLogger.Data)
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger()

	assert.NotNil(t, logger)
	assert.IsType(t, &logrus.Entry{}, logger)
	assert.Empty(t, logger.Data)
}

func TestLoggerChaining(t *testing.T) {
	// Test that we can chain logger operations
	logger1 := NewLogger().WithField("step", 1)
	ctx := WithLogger(context.Background(), logger1)

	logger2 := GetLogger(ctx).WithField("step", 2)
	ctx = WithLogger(ctx, logger2)

	finalLogger := GetLogger(ctx)

	assert.NotNil(t, finalLogger)
	assert.Equal(t, 2, finalLogger.Data["step"])
}

func TestLoggerImmutability(t *testing.T) {
	// Test that modifying retrieved logger doesn't affect the original
	originalLogger := logrus.NewEntry(logrus.New()).WithField("original", "value")
	ctx := WithLogger(context.Background(), originalLogger)

	retrievedLogger := GetLogger(ctx)
	modifiedLogger := retrievedLogger.WithField("modified", "new_value")

	// Original logger should remain unchanged
	assert.Equal(t, "value", originalLogger.Data["original"])
	assert.NotContains(t, originalLogger.Data, "modified")

	// Modified logger should have both fields
	assert.Equal(t, "value", modifiedLogger.Data["original"])
	assert.Equal(t, "new_value", modifiedLogger.Data["modified"])
}

func TestContextIsolation(t *testing.T) {
	// Test that different contexts maintain separate loggers
	logger1 := logrus.NewEntry(logrus.New()).WithField("context", "first")
	logger2 := logrus.NewEntry(logrus.New()).WithField("context", "second")

	ctx1 := WithLogger(context.Background(), logger1)
	ctx2 := WithLogger(context.Background(), logger2)

	retrievedLogger1 := GetLogger(ctx1)
	retrievedLogger2 := GetLogger(ctx2)

	assert.Equal(t, "first", retrievedLogger1.Data["context"])
	assert.Equal(t, "second", retrievedLogger2.Data["context"])
	assert.NotEqual(t, retrievedLogger1, retrievedLogger2)
}
