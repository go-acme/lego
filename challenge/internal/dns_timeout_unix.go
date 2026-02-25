//go:build !windows

package internal

import "time"

// DNSTimeout is used as the default DNS timeout on Unix-like systems.
const DNSTimeout = 10 * time.Second
