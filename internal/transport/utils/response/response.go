package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lzimin05/course-todo/internal/transport/dto/utils"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

func SendError(ctx context.Context, w http.ResponseWriter, status int, message string) {
	const op = "response.SendError"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp, err := json.Marshal(dto.ErrorResponse{Message: message})
	if err != nil {
		logctx.GetLogger(ctx).Error("failed to marshal response: ", err.Error())
		return
	}

	if _, err := w.Write(resp); err != nil {
		logctx.GetLogger(ctx).Error("failed to write response: ", err.Error())
	}
}

func SendJSONResponse(ctx context.Context, w http.ResponseWriter, statusCode int, body any) {
	const op = "response.SendJSONResponse"
	if body == nil {
		w.WriteHeader(statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logctx.GetLogger(ctx).Error("failed to marshal response", err.Error())
		return
	}

	w.WriteHeader(statusCode)
	if _, err := w.Write(resp); err != nil {
		logctx.GetLogger(ctx).Error("failed to write response", err.Error())
	}
}
