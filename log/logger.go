package log

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	defaultLogger.Store(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

// Default returns the default [Logger].
func Default() *slog.Logger { return defaultLogger.Load() }

// SetDefault makes l the default [Logger], which is used by
// the top-level functions [Info], [Debug] and so on.
func SetDefault(l *slog.Logger) {
	defaultLogger.Store(l)
}

// Fatal calls [Logger.Error] on the default logger and exit with code 1.
func Fatal(msg string, args ...slog.Attr) {
	Error(msg, args...)
	os.Exit(1)
}

// Debug calls [Logger.Debug] on the default logger.
func Debug(msg string, args ...slog.Attr) {
	Default().LogAttrs(context.Background(), slog.LevelDebug, msg, args...)
}

// Info calls [Logger.Info] on the default logger.
func Info(msg string, args ...slog.Attr) {
	Default().LogAttrs(context.Background(), slog.LevelInfo, msg, args...)
}

// Warn calls [Logger.Warn] on the default logger.
func Warn(msg string, args ...slog.Attr) {
	Default().LogAttrs(context.Background(), slog.LevelWarn, msg, args...)
}

// Error calls [Logger.Error] on the default logger.
func Error(msg string, args ...slog.Attr) {
	Default().LogAttrs(context.Background(), slog.LevelError, msg, args...)
}
