package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_launchHook_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		hook     string
		timeout  time.Duration
		expected string
	}{
		{
			desc:     "kill the hook",
			hook:     "sleep 5",
			timeout:  1 * time.Second,
			expected: "wait command: signal: killed",
		},
		{
			desc:     "context timeout",
			hook:     "echo foo",
			timeout:  1 * time.Nanosecond,
			expected: "start command: context deadline exceeded",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := launchHook(test.hook, test.timeout, map[string]string{})
			require.EqualError(t, err, test.expected)
		})
	}
}
