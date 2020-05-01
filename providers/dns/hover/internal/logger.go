package internal

import (
	"log"
)

// YALI -- Yet Another Logger Interface -- reduces the logger facility needed to as few functions
// as possible to allow others to be slotted in.
type YALI interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// NopLogger reduces spin while not logging
// https://gist.github.com/Avinash-Bhat/48c4f06b0cc840d9fd6c#file-log_test-go
//
// Intended to be a compatible implementation for YALI for low-cost log discarding
type NopLogger struct {
	*log.Logger
}

// Printf offers a relatively efficient discarding function for log messages when not logging
func (l *NopLogger) Printf(format string, v ...interface{}) {
	// noop
}

// Println offers a relatively efficient discarding function for log messages when not logging
func (l *NopLogger) Println(v ...interface{}) {
	// noop
}
