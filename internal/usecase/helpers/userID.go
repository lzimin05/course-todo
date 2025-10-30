package helpers

import (
	"context"

	"github.com/google/uuid"
	"github.com/lzimin05/course-todo/internal/models/domains"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDStr, isExist := ctx.Value(domains.UserIDKey{}).(string)
	if !isExist {
		return uuid.Nil, errs.NewNotFoundError("user not found")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errs.ErrInvalidID
	}

	return userID, nil
}
