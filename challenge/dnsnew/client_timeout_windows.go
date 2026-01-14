//go:build windows

package dnsnew

import "time"

// dnsTimeout is used to override the default DNS timeout of 20 seconds.
const dnsTimeout = 20 * time.Second
