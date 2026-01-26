package log

import (
	"log"
	"os"
	"strings"
)

// Logger is an optional custom logger.
// Users can assign their own logger (e.g., logrus) to get structured logging.
//
// Example with logrus JSON output and structured fields:
//
//	import (
//		"github.com/go-acme/lego/v4/log"
//		"github.com/sirupsen/logrus"
//	)
//
//	type LogrusWrapper struct {
//		*logrus.Logger
//	}
//
//	func (l *LogrusWrapper) WithFields(fields map[string]any) log.FieldEntry {
//		return l.Logger.WithFields(logrus.Fields(fields))
//	}
//
//	logger := logrus.New()
//	logger.SetFormatter(&logrus.JSONFormatter{})
//	log.Logger = &LogrusWrapper{Logger: logger}
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

// LevelLogger extends StdLogger with level-specific methods.
type LevelLogger interface {
	StdLogger
	Warn(args ...any)
	Warnf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
}

// FieldLogger is an optional interface for structured logging.
// Loggers like logrus implement this interface.
// The returned value must have Info(args ...any) and Warn(args ...any) methods.
type FieldLogger interface {
	WithFields(fields map[string]any) FieldEntry
}

// FieldEntry is the interface for the object returned by WithFields.
type FieldEntry interface {
	Info(args ...any)
	Warn(args ...any)
}

// Fatal writes a log entry.
func Fatal(args ...any) {
	Logger.Fatal(args...)
}

// Fatalf writes a log entry.
func Fatalf(format string, args ...any) {
	Logger.Fatalf(format, args...)
}

// Print writes a log entry.
func Print(args ...any) {
	Logger.Print(args...)
}

// Println writes a log entry.
func Println(args ...any) {
	Logger.Println(args...)
}

// Printf writes a log entry.
func Printf(format string, args ...any) {
	Logger.Printf(format, args...)
}

// Warnf writes a warning log entry.
func Warnf(format string, args ...any) {
	if levelLogger, ok := Logger.(LevelLogger); ok {
		levelLogger.Warnf(format, args...)
	} else {
		Printf("[WARN] "+format, args...)
	}
}

// Infof writes an info log entry.
func Infof(format string, args ...any) {
	if levelLogger, ok := Logger.(LevelLogger); ok {
		levelLogger.Infof(format, args...)
	} else {
		Printf("[INFO] "+format, args...)
	}
}

// Info writes an info log entry with structured fields.
// When the logger implements FieldLogger (e.g., logrus), fields are logged
// as structured data. Otherwise, fields are appended to the message.
//
// Example:
//
//	log.Info("acme: Obtaining certificate", "domain", "example.com")
//	log.Info("acme: Obtaining certificate", "domain", "example.com", "domains", []string{"example.com", "www.example.com"})
func Info(msg string, keyvals ...any) {
	fields := keyValsToMap(keyvals)

	if len(fields) > 0 {
		if fl, ok := Logger.(FieldLogger); ok {
			fl.WithFields(fields).Info(msg)
			return
		}
	}

	if len(fields) == 0 {
		Infof("%s", msg)
	} else {
		Infof("%s %s", msg, formatFields(fields))
	}
}

// Warn writes a warning log entry with structured fields.
func Warn(msg string, keyvals ...any) {
	fields := keyValsToMap(keyvals)

	if len(fields) > 0 {
		if fl, ok := Logger.(FieldLogger); ok {
			fl.WithFields(fields).Warn(msg)
			return
		}
	}

	if len(fields) == 0 {
		Warnf("%s", msg)
	} else {
		Warnf("%s %s", msg, formatFields(fields))
	}
}

// keyValsToMap converts key-value pairs to a map.
func keyValsToMap(keyvals []any) map[string]any {
	fields := make(map[string]any)
	for i := 0; i+1 < len(keyvals); i += 2 {
		if key, ok := keyvals[i].(string); ok {
			fields[key] = keyvals[i+1]
		}
	}
	return fields
}

// formatFields formats fields for fallback text logging.
func formatFields(fields map[string]any) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for k, v := range fields {
		parts = append(parts, formatKeyValue(k, v))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// formatKeyValue formats a single key-value pair.
func formatKeyValue(key string, value any) string {
	switch v := value.(type) {
	case string:
		return key + "=" + v
	case []string:
		return key + "=[" + strings.Join(v, ", ") + "]"
	default:
		return key + "=" + stringify(v)
	}
}

// stringify converts a value to string.
func stringify(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if stringer, ok := v.(interface{ String() string }); ok {
		return stringer.String()
	}
	return ""
}
