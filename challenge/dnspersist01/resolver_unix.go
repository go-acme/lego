//go:build !windows

package dnspersist01

import "time"

// dnsTimeout is used as the default DNS timeout on Unix-like systems.
const dnsTimeout = 10 * time.Second
