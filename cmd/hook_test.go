package cmd

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_launchHook(t *testing.T) {
	err := launchHook("echo foo", 1*time.Second, map[string]string{})
	require.NoError(t, err)
}

func Test_launchHook_errors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows")
	}

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
			expected: "hook timed out",
		},
		{
			desc:     "context timeout on Start",
			hook:     "echo foo",
			timeout:  1 * time.Nanosecond,
			expected: "start command: context deadline exceeded",
		},
		{
			desc:     "multiple short sleeps",
			hook:     "./testdata/sleepy.sh",
			timeout:  1 * time.Second,
			expected: "hook timed out",
		},
		{
			desc:     "long sleep",
			hook:     "./testdata/sleeping_beauty.sh",
			timeout:  1 * time.Second,
			expected: "hook timed out",
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
