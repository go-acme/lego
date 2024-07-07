package log

import (
	"log"
	"os"
)

// Logger is an optional custom logger.
var Logger StdLogger = log.New(os.Stderr, "", log.LstdFlags)

// StdLogger interface for Standard Logger.
type StdLogger interface {
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})
	Print(args ...interface{})
	Println(args ...interface{})
	Printf(format string, args ...interface{})
}

var Quiet = false

// Fatal writes a log entry.
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf writes a log entry.
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// Print writes a log entry.
func Print(args ...interface{}) {
	if !Quiet {
		Logger.Print(args...)
	}
}

// Println writes a log entry.
func Println(args ...interface{}) {
	if !Quiet {
		Logger.Println(args...)
	}
}

// Printf writes a log entry.
func Printf(format string, args ...interface{}) {
	if !Quiet {
		Logger.Printf(format, args...)
	}
}

// Warnf writes a log entry.
func Warnf(format string, args ...interface{}) {
	Printf("[WARN] "+format, args...)
}

// Infof writes a log entry.
func Infof(format string, args ...interface{}) {
	if !Quiet {
		Printf("[INFO] "+format, args...)
	}
}
