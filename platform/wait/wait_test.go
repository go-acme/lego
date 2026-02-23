package wait

import (
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFor_timeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		var io atomic.Int64

		err := For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)

			return false, nil
		})

		assert.Equal(t, 3*time.Second, time.Since(now))
		require.EqualValues(t, 3, io.Load())
		require.EqualError(t, err, "test: time limit exceeded")
	})
}

func TestFor_timeout_with_error(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		var io atomic.Int64

		err := For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)

			// This allows be sure that the latest previous error is returned.
			if io.Load() == 1 {
				return false, errors.New("oops")
			}

			return false, nil
		})

		assert.Equal(t, 3*time.Second, time.Since(now))
		require.EqualValues(t, 3, io.Load())
		require.EqualError(t, err, "test: time limit exceeded: last error: oops")
	})
}

func TestFor_stop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		var io atomic.Int64

		err := For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)

			return true, nil
		})

		assert.Equal(t, 0*time.Second, time.Since(now))
		require.NoError(t, err)
		require.EqualValues(t, 1, io.Load())
	})
}

func TestFor_stop_with_error(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		now := time.Now()

		var io atomic.Int64

		err := For("test", 3*time.Second, 1*time.Second, func() (bool, error) {
			io.Add(1)

			return true, errors.New("oops")
		})

		assert.Equal(t, 0*time.Second, time.Since(now))
		require.EqualError(t, err, "oops")
		require.EqualValues(t, 1, io.Load())
	})
}
