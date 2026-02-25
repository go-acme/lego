//go:build !windows

package dnspersist01

/*
 * NOTE(ldez): This file is a duplication of `dns01/client_timeout_unix.go`.
 * The 2 files should be kept in sync.
 */

import "time"

// dnsTimeout is used as the default DNS timeout on Unix-like systems.
const dnsTimeout = 10 * time.Second
