package wait

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v5/log"
)

// For polls the given function 'f', once every 'interval', up to 'timeout'.
func For(timeout, interval time.Duration, f func() (bool, error)) error {
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
			return err
		}

		if err != nil {
			log.Debug("Waiting for condition failed.", log.ErrorAttr(err))

			lastErr = err
		}

		time.Sleep(interval)
	}
}

// Retry retries the given operation until it succeeds or the context is canceled.
// Similar to [backoff.Retry] but with a different signature.
func Retry(ctx context.Context, operation func() error, opts ...backoff.RetryOption) error {
	_, err := backoff.Retry(ctx, func() (any, error) {
		return nil, operation()
	}, opts...)

	return err
}

func SimpleNotify(message string) backoff.Notify {
	return func(err error, duration time.Duration) {
		log.Debug(message, log.ErrorAttr(err), log.DurationAttr("duration", duration))
	}
}
