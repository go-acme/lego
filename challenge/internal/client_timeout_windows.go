//go:build windows

package internal

import "time"

// dnsTimeout is used as the default DNS timeout on Windows.
const dnsTimeout = 20 * time.Second
