package wait

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TODO(ldez): rewrite those tests when upgrading to go1.25 as minimum Go version.

func TestFor_timeout(t *testing.T) {
	var io atomic.Int64

	c := make(chan error)

	go func() {
		c <- For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)
			if io.Load() == 1 {
				return false, nil
			}

			return false, nil
		})
	}()

	timeout := time.After(6 * time.Second)

	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		require.EqualError(t, err, "test: time limit exceeded")
	}

	require.EqualValues(t, 3, io.Load())
}

func TestFor_timeout_with_error(t *testing.T) {
	var io atomic.Int64

	c := make(chan error)

	go func() {
		c <- For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)

			// This allows be sure that the latest previous error is returned.
			if io.Load() == 1 {
				return false, errors.New("oops")
			}

			return false, nil
		})
	}()

	timeout := time.After(6 * time.Second)

	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		require.EqualError(t, err, "test: time limit exceeded: last error: oops")
	}

	require.EqualValues(t, 3, io.Load())
}

func TestFor_stop(t *testing.T) {
	var io atomic.Int64

	c := make(chan error)

	go func() {
		c <- For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)

			return true, nil
		})
	}()

	timeout := time.After(6 * time.Second)

	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		require.NoError(t, err)
	}

	require.EqualValues(t, 1, io.Load())
}

func TestFor_stop_with_error(t *testing.T) {
	var io atomic.Int64

	c := make(chan error)

	go func() {
		c <- For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)

			return true, errors.New("oops")
		})
	}()

	timeout := time.After(6 * time.Second)

	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		require.EqualError(t, err, "oops")
	}

	require.EqualValues(t, 1, io.Load())
}
