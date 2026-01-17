package exec

import (
	"context"
	"log/slog"

	"github.com/stretchr/testify/mock"
)

type LogHandler struct {
	mock.Mock
}

func (l *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (l *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	l.Called(ctx, record)

	return nil
}

func (l *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("implement me")
}

func (l *LogHandler) WithGroup(name string) slog.Handler {
	panic("implement me")
}
