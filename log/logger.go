package log

import (
	"log"
	"os"
)

var (
	// Logger is an optional custom logger.
	Logger *log.Logger
)

// Fatalf writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Fatalf(format string, args ...interface{}) {
	if Logger == nil {
		Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	Logger.Fatalf(format, args...)
}

// Printf writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Printf(format string, args ...interface{}) {
	if Logger == nil {
		Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	Logger.Printf(format, args...)
}
