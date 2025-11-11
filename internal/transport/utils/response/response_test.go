package response

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	dto "github.com/lzimin05/course-todo/internal/transport/dto/utils"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func setupResponseTest() context.Context {
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.PanicLevel) // Suppress log output during tests
	return logctx.WithLogger(context.Background(), logger)
}

func TestSendError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		expectedStatus int
		expectedBody   dto.ErrorResponse
	}{
		{
			name:           "bad request error",
			status:         400,
			message:        "Invalid input",
			expectedStatus: 400,
			expectedBody:   dto.ErrorResponse{Message: "Invalid input"},
		},
		{
			name:           "unauthorized error",
			status:         401,
			message:        "Unauthorized access",
			expectedStatus: 401,
			expectedBody:   dto.ErrorResponse{Message: "Unauthorized access"},
		},
		{
			name:           "internal server error",
			status:         500,
			message:        "Internal server error",
			expectedStatus: 500,
			expectedBody:   dto.ErrorResponse{Message: "Internal server error"},
		},
		{
			name:           "empty message",
			status:         404,
			message:        "",
			expectedStatus: 404,
			expectedBody:   dto.ErrorResponse{Message: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupResponseTest()
			w := httptest.NewRecorder()

			SendError(ctx, w, tt.status, tt.message)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response dto.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestSendJSONResponse_WithBody(t *testing.T) {
	type TestData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name           string
		statusCode     int
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "successful response with struct",
			statusCode:     200,
			body:           TestData{ID: 1, Name: "test"},
			expectedStatus: 200,
		},
		{
			name:           "created response with map",
			statusCode:     201,
			body:           map[string]string{"message": "created"},
			expectedStatus: 201,
		},
		{
			name:           "response with slice",
			statusCode:     200,
			body:           []string{"item1", "item2", "item3"},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupResponseTest()
			w := httptest.NewRecorder()

			SendJSONResponse(ctx, w, tt.statusCode, tt.body)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.NotEmpty(t, w.Body.Bytes())

			// Verify that the response can be unmarshaled back
			var response interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
		})
	}
}

func TestSendJSONResponse_WithNilBody(t *testing.T) {
	ctx := setupResponseTest()
	w := httptest.NewRecorder()

	SendJSONResponse(ctx, w, 204, nil)

	assert.Equal(t, 204, w.Code)
	assert.Empty(t, w.Body.Bytes())
	assert.Empty(t, w.Header().Get("Content-Type"))
}

func TestSendJSONResponse_WithUnmarshalableBody(t *testing.T) {
	ctx := setupResponseTest()
	w := httptest.NewRecorder()

	// Create an unmarshalable body (channel cannot be marshaled to JSON)
	unmarshalableBody := make(chan int)

	SendJSONResponse(ctx, w, 200, unmarshalableBody)

	assert.Equal(t, 500, w.Code)
	assert.Empty(t, w.Body.Bytes())
}

func TestSendError_SpecialCharacters(t *testing.T) {
	ctx := setupResponseTest()
	w := httptest.NewRecorder()

	message := "Error with special chars: <>&\"'"
	SendError(ctx, w, 400, message)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, message, response.Message)
}

func TestSendJSONResponse_EmptyStructs(t *testing.T) {
	type EmptyStruct struct{}

	tests := []struct {
		name string
		body interface{}
	}{
		{
			name: "empty struct",
			body: EmptyStruct{},
		},
		{
			name: "empty map",
			body: map[string]interface{}{},
		},
		{
			name: "empty slice",
			body: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupResponseTest()
			w := httptest.NewRecorder()

			SendJSONResponse(ctx, w, 200, tt.body)

			assert.Equal(t, 200, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.NotEmpty(t, w.Body.Bytes())

			// Verify valid JSON
			var response interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
		})
	}
}
