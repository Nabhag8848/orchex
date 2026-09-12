package logger

import (
	"log/slog"
	"os"
	"strings"
)

func Configure(level string) {
	logLevel := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))
}

func Fatal(message string, args ...any) {
	slog.Error(message, args...)
	os.Exit(1)
}
