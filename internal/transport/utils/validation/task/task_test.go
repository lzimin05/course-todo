package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidationTask_Success(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	err := ValidationTask("Valid Task", 2, future)
	assert.NoError(t, err)
}

func TestValidationTask_Title_Validation(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		title       string
		expectedErr string
	}{
		{
			name:        "empty title",
			title:       "",
			expectedErr: "title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationTask(tt.title, 2, future)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidationTask_Importance_Validation(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		importance  int
		expectedErr string
	}{
		{
			name:        "importance too low",
			importance:  0,
			expectedErr: "importance must be between 1 and 3",
		},
		{
			name:        "importance too high",
			importance:  4,
			expectedErr: "importance must be between 1 and 3",
		},
		{
			name:        "negative importance",
			importance:  -1,
			expectedErr: "importance must be between 1 and 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationTask("Valid Task", tt.importance, future)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidationTask_Deadline_Validation(t *testing.T) {
	tests := []struct {
		name        string
		deadline    time.Time
		expectedErr string
	}{
		{
			name:        "deadline in the past",
			deadline:    time.Now().Add(-24 * time.Hour),
			expectedErr: "deadline must be in the future",
		},
		{
			name:        "deadline exactly now (should fail due to processing time)",
			deadline:    time.Now(),
			expectedErr: "deadline must be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationTask("Valid Task", 2, tt.deadline)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidationTask_ValidCases(t *testing.T) {
	validCases := []struct {
		name       string
		title      string
		importance int
		deadline   time.Time
	}{
		{
			name:       "importance level 1",
			title:      "Low priority task",
			importance: 1,
			deadline:   time.Now().Add(time.Hour),
		},
		{
			name:       "importance level 2",
			title:      "Medium priority task",
			importance: 2,
			deadline:   time.Now().Add(24 * time.Hour),
		},
		{
			name:       "importance level 3",
			title:      "High priority task",
			importance: 3,
			deadline:   time.Now().Add(7 * 24 * time.Hour),
		},
		{
			name:       "long title",
			title:      "This is a very long task title that should still be valid",
			importance: 2,
			deadline:   time.Now().Add(48 * time.Hour),
		},
		{
			name:       "far future deadline",
			title:      "Future task",
			importance: 1,
			deadline:   time.Now().Add(365 * 24 * time.Hour),
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidationTask(tc.title, tc.importance, tc.deadline)
			assert.NoError(t, err)
		})
	}
}

func TestValidationTask_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		importance int
		deadline   time.Time
		shouldFail bool
	}{
		{
			name:       "minimum valid importance",
			title:      "Valid Task",
			importance: 1,
			deadline:   time.Now().Add(time.Minute),
			shouldFail: false,
		},
		{
			name:       "maximum valid importance",
			title:      "Valid Task",
			importance: 3,
			deadline:   time.Now().Add(time.Minute),
			shouldFail: false,
		},
		{
			name:       "very short future deadline",
			title:      "Valid Task",
			importance: 2,
			deadline:   time.Now().Add(time.Second),
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationTask(tt.title, tt.importance, tt.deadline)
			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
