//go:build !windows

package dns01

import "time"

// dnsTimeout is used to override the default DNS timeout of 10 seconds.
var dnsTimeout = 10 * time.Second
