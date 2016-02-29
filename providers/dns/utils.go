package dns

import (
	"fmt"
	"time"
)

// ToFqdn converts the name into a fqdn appending a trailing dot.
func ToFqdn(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	}
	return name + "."
}

// UnFqdn converts the fqdn into a name removing the trailing dot.
func UnFqdn(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}

// WaitFor polls the given function 'f', once every 'interval' seconds, up to 'timeout' seconds.
func WaitFor(timeout, interval int, f func() (bool, error)) error {
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
