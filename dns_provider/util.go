package dns_provider

import (
	"fmt"
	"log"
	"time"

	"github.com/xenolf/lego/acme"
)

// toFqdn converts the name into a fqdn appending a trailing dot.
func toFqdn(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	}
	return name + "."
}

// unFqdn converts the fqdn into a name removing the trailing dot.
func unFqdn(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}

// waitFor polls the given function 'f', once every 'interval' seconds, up to 'timeout' seconds.
func waitFor(timeout, interval int, f func() (bool, error)) error {
	var lastErr string
	timeup := time.After(time.Duration(timeout) * time.Second)
	for {
		select {
		case <-timeup:
			return fmt.Errorf("Time limit exceeded. Last error: %s", lastErr)
		default:
		}

		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			lastErr = err.Error()
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// logf writes a log entry. It uses acme.Logger if not
// nil, otherwise it uses the default log.Logger.
func logf(format string, args ...interface{}) {
	if acme.Logger != nil {
		acme.Logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
