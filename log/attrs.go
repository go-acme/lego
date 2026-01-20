package log

import (
	"log/slog"
)

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
