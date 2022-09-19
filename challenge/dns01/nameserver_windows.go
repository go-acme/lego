//go:build windows

package dns01

import "time"

// dnsTimeout is used to override the default DNS timeout of 20 seconds.
var dnsTimeout = 20 * time.Second
