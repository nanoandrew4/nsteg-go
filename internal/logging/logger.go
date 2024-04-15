package logging

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func BuildLogger() *Logger {
	logger := Logger{Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	return &logger
}

func BuildLoggerFromCtx(ctx *gin.Context) *Logger {
	logger := Logger{Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	logger = Logger{Logger: logger.With("path", ctx.Request.URL.Path)}
	return &logger
}

func (l *Logger) WithError(err error) *Logger {
	modifiedLogger := Logger{Logger: l.With("error", err.Error())}
	return &modifiedLogger
}
