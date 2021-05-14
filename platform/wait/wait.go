package wait

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/log"
)

// For polls the given function 'f', once every 'interval', up to 'timeout'.
func For(msg string, timeout, interval time.Duration, f func() (bool, error)) error {
	log.Infof("Wait for %s [timeout: %s, interval: %s]", msg, timeout, interval)

	var lastErr error
	timeUp := time.After(timeout)
	for {
		select {
		case <-timeUp:
			if lastErr == nil {
				return errors.New("time limit exceeded")
			}
			return fmt.Errorf("time limit exceeded: last error: %w", lastErr)
		default:
		}

		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			lastErr = err
		}

		time.Sleep(interval)
	}
}
