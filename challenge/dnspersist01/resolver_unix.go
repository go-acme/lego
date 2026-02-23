//go:build !windows

package dnspersist01

import "time"

// defaultDNSTimeout is used as the default DNS timeout on Unix-like systems.
const defaultDNSTimeout = 10 * time.Second

// DefaultDNSTimeout returns the default DNS timeout.
func DefaultDNSTimeout() time.Duration {
	return defaultDNSTimeout
}
