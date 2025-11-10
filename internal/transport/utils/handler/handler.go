package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/lzimin05/course-todo/internal/models/errs"
	"github.com/lzimin05/course-todo/internal/transport/utils/response"
)

func HandleError(ctx context.Context, w http.ResponseWriter, err error, defaultMsg string) {
	switch {
	case errors.Is(err, errs.ErrNoAccess):
		response.SendError(ctx, w, http.StatusForbidden, "Access denied")
	case errors.Is(err, errs.ErrNotOwner):
		response.SendError(ctx, w, http.StatusForbidden, "Insufficient permissions")
	case errors.Is(err, errs.ErrNotFound):
		response.SendError(ctx, w, http.StatusNotFound, "Resource not found")
	case errors.Is(err, errs.ErrInvalidID):
		response.SendError(ctx, w, http.StatusBadRequest, "Invalid ID format")
	default:
		response.SendError(ctx, w, http.StatusInternalServerError, defaultMsg)
	}
}
