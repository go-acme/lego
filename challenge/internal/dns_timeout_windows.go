//go:build windows

package internal

import "time"

// DNSTimeout is used as the default DNS timeout on Windows.
const DNSTimeout = 20 * time.Second
