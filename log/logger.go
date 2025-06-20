package log

import (
	"log"
	"os"
)

// Logger is an optional custom logger.
var Logger StdLogger = log.New(os.Stderr, "", log.LstdFlags)

// StdLogger interface for Standard Logger.
type StdLogger interface {
	Fatal(args ...any)
	Fatalln(args ...any)
	Fatalf(format string, args ...any)
	Print(args ...any)
	Println(args ...any)
	Printf(format string, args ...any)
}

// Fatal writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Fatal(args ...any) {
	Logger.Fatal(args...)
}

// Fatalf writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Fatalf(format string, args ...any) {
	Logger.Fatalf(format, args...)
}

// Print writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Print(args ...any) {
	Logger.Print(args...)
}

// Println writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Println(args ...any) {
	Logger.Println(args...)
}

// Printf writes a log entry.
// It uses Logger if not nil, otherwise it uses the default log.Logger.
func Printf(format string, args ...any) {
	Logger.Printf(format, args...)
}

// Warnf writes a log entry.
func Warnf(format string, args ...any) {
	Printf("[WARN] "+format, args...)
}

// Infof writes a log entry.
func Infof(format string, args ...any) {
	Printf("[INFO] "+format, args...)
}
