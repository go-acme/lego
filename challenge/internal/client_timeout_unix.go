//go:build !windows

package internal

import "time"

// dnsTimeout is used as the default DNS timeout on Unix-like systems.
const dnsTimeout = 10 * time.Second
