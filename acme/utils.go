package acme

import (
	"fmt"
	"time"
)

// WaitFor polls the given function 'f', once every 'interval' seconds, up to 'timeout' seconds.
func WaitFor(timeout, interval time.Duration, f func() (bool, error)) error {
	var lastErr string
	timeup := time.After(timeout * time.Second)
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

		time.Sleep(interval * time.Second)
	}
}
