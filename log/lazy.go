package log

import (
	"context"
	"fmt"
	"log/slog"
)

type LazyMessage struct {
	msg  string
	args []any
}

func LazySprintf(msg string, args ...any) LazyMessage {
	return LazyMessage{
		msg:  msg,
		args: args,
	}
}

func (l LazyMessage) String() string {
	return fmt.Sprintf(l.msg, l.args...)
}

// Debugf calls [Logger.Debug] on the default logger.
func Debugf(msg LazyMessage, args ...slog.Attr) {
	logLazy(slog.LevelDebug, msg, args...)
}

// Infof calls [Logger.Info] on the default logger.
func Infof(msg LazyMessage, args ...slog.Attr) {
	logLazy(slog.LevelInfo, msg, args...)
}

// Warnf calls [Logger.Warn] on the default logger.
func Warnf(msg LazyMessage, args ...slog.Attr) {
	logLazy(slog.LevelWarn, msg, args...)
}

// Errorf calls [Logger.Error] on the default logger.
func Errorf(msg LazyMessage, args ...slog.Attr) {
	logLazy(slog.LevelError, msg, args...)
}

func logLazy(level slog.Level, msg LazyMessage, args ...slog.Attr) {
	ctx := context.Background()

	if Default().Enabled(ctx, level) {
		Default().LogAttrs(ctx, level, msg.String(), args...)
	}
}
