package log

import (
	"log"
	"os"
)

// Logger is an optional custom logger to stdout.
// ErrLogger is an optional custom logger to stderr.
var Logger StdLogger = log.New(os.Stdout, "", log.LstdFlags)
var ErrLogger StdLogger = log.New(os.Stderr, "", log.LstdFlags)

// StdLogger interface.
type StdLogger interface {
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})
	Print(args ...interface{})
	Println(args ...interface{})
	Printf(format string, args ...interface{})
}

// Fatal writes a log entry.
// It uses ErrLogger if not nil, otherwise it uses the default log.Logger.
func Fatal(args ...interface{}) {
	ErrLogger.Fatal(args...)
}

// Fatalf writes a log entry.
// It uses ErrLogger if not nil, otherwise it uses the default log.Logger.
func Fatalf(format string, args ...interface{}) {
	ErrLogger.Fatalf(format, args...)
}

// Print writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Print(args ...interface{}) {
	Logger.Print(args...)
}

// Println writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Println(args ...interface{}) {
	Logger.Println(args...)
}

// Printf writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Printf(format string, args ...interface{}) {
	Logger.Printf(format, args...)
}

// Warnf writes a log entry.
// It uses ErrLogger if not nil, otherwise it uses the default log.Logger.
func Warnf(format string, args ...interface{}) {
	ErrLogger.Printf("[WARN] "+format, args...)
}

// Infof writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Infof(format string, args ...interface{}) {
	Logger.Printf("[INFO] "+format, args...)
}
