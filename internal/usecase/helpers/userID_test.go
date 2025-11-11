package helpers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/internal/models/domains"
)

func TestGetUserIDFromContext_Success(t *testing.T) {
	expectedID := uuid.New()
	ctx := context.WithValue(context.Background(), domains.UserIDKey{}, expectedID.String())

	result, err := GetUserIDFromContext(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, result)
}

func TestGetUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	result, err := GetUserIDFromContext(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Equal(t, uuid.Nil, result)
}

func TestGetUserIDFromContext_InvalidFormat(t *testing.T) {
	ctx := context.WithValue(context.Background(), domains.UserIDKey{}, "invalid-uuid")

	result, err := GetUserIDFromContext(ctx)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
}
