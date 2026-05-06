package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormattableDuration(t *testing.T) {
	testCases := []struct {
		desc     string
		date     time.Time
		duration time.Duration
		expected string
	}{
		{
			desc:     "all",
			duration: 47*time.Hour + 3*time.Minute + 8*time.Second + 1234567890*time.Nanosecond,
			expected: "1d23h3m9s234567890ns",
		},
		{
			desc:     "without nanoseconds",
			duration: 47*time.Hour + 3*time.Minute + 8*time.Second,
			expected: "1d23h3m8s",
		},
		{
			desc:     "without seconds",
			duration: 47*time.Hour + 3*time.Minute + 2*time.Nanosecond,
			expected: "1d23h3m2ns",
		},
		{
			desc:     "without minutes",
			duration: 47*time.Hour + 8*time.Second + 2*time.Nanosecond,
			expected: "1d23h8s2ns",
		},
		{
			desc:     "without hours",
			duration: 3*time.Minute + 8*time.Second + 2*time.Nanosecond,
			expected: "3m8s2ns",
		},
		{
			desc:     "only hours",
			duration: 23 * time.Hour,
			expected: "23h",
		},
		{
			desc:     "only days",
			duration: 48 * time.Hour,
			expected: "2d",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, FormattableDuration(test.duration).String())
		})
	}
}
