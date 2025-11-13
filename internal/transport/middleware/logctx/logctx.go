package logctx

import (
	"context"

	"github.com/lzimin05/course-todo/internal/models/domains"

	"github.com/sirupsen/logrus"
)

func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, domains.LoggerKey{}, logger)
}

func GetLogger(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(domains.LoggerKey{}).(*logrus.Entry); ok {
		return logger
	}

	return logrus.NewEntry(logrus.New())
}

func NewLogger() *logrus.Entry {
	return logrus.NewEntry(logrus.New())
}
