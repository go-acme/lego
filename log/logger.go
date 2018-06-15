package log

import (
	"log"
	"os"
)

// Logger is an optional custom logger.
var Logger *log.Logger

func initDefaultLogger() {
	if Logger == nil {
		Logger = log.New(os.Stdout, "", log.LstdFlags)
	}
}

// Fatal writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Fatal(args ...interface{}) {
	initDefaultLogger()

	Logger.Fatal(args...)
}

// Fatalf writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Fatalf(format string, args ...interface{}) {
	initDefaultLogger()

	Logger.Fatalf(format, args...)
}

// Print writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Print(args ...interface{}) {
	initDefaultLogger()

	Logger.Print(args...)
}

// Println writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Println(args ...interface{}) {
	initDefaultLogger()

	Logger.Println(args...)
}

// Printf writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Printf(format string, args ...interface{}) {
	initDefaultLogger()

	Logger.Printf(format, args...)
}

// Warnf writes a log entry.
func Warnf(format string, args ...interface{}) {
	Printf("[WARN] "+format, args...)
}

// Infof writes a log entry.
func Infof(format string, args ...interface{}) {
	Printf("[INFO] "+format, args...)
}
