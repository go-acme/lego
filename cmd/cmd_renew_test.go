package cmd

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_merge(t *testing.T) {
	testCases := []struct {
		desc        string
		prevDomains []string
		nextDomains []string
		expected    []string
	}{
		{
			desc:        "all empty",
			prevDomains: []string{},
			nextDomains: []string{},
			expected:    []string{},
		},
		{
			desc:        "next empty",
			prevDomains: []string{"a", "b", "c"},
			nextDomains: []string{},
			expected:    []string{"a", "b", "c"},
		},
		{
			desc:        "prev empty",
			prevDomains: []string{},
			nextDomains: []string{"a", "b", "c"},
			expected:    []string{"a", "b", "c"},
		},
		{
			desc:        "merge append",
			prevDomains: []string{"a", "b", "c"},
			nextDomains: []string{"a", "c", "d"},
			expected:    []string{"a", "b", "c", "d"},
		},
		{
			desc:        "merge same",
			prevDomains: []string{"a", "b", "c"},
			nextDomains: []string{"a", "b", "c"},
			expected:    []string{"a", "b", "c"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := merge(test.prevDomains, test.nextDomains)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_isInRenewalPeriod_days(t *testing.T) {
	now := time.Date(2025, 1, 19, 1, 1, 1, 1, time.UTC)

	oneDay := 24 * time.Hour

	testCases := []struct {
		desc     string
		cert     *x509.Certificate
		days     int
		expected assert.BoolAssertionFunc
	}{
		{
			desc: "days: 30 days, NotAfter now",
			cert: &x509.Certificate{
				NotAfter: now,
			},
			days:     30,
			expected: assert.True,
		},
		{
			desc: "days: 30 days, NotAfter 31 days",
			cert: &x509.Certificate{
				NotAfter: now.Add(31*oneDay + 1*time.Second),
			},
			days:     30,
			expected: assert.False,
		},
		{
			desc: "days: 30 days, NotAfter 30 days",
			cert: &x509.Certificate{
				NotAfter: now.Add(30 * oneDay),
			},
			days:     30,
			expected: assert.True,
		},
		{
			desc: "days: 0 days, NotAfter 30 days: only the day of the expiration",
			cert: &x509.Certificate{
				NotAfter: now.Add(30 * oneDay),
			},
			days:     0,
			expected: assert.False,
		},
		{
			desc: "days: -1 days, NotAfter 30 days: always renew",
			cert: &x509.Certificate{
				NotAfter: now.Add(30 * oneDay),
			},
			days:     -1,
			expected: assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			actual := isInRenewalPeriod(test.cert, "foo.com", test.days, now)

			test.expected(t, actual)
		})
	}
}

func Test_isInRenewalPeriod_dynamic(t *testing.T) {
	testCases := []struct {
		desc                string
		now                 time.Time
		notBefore, notAfter time.Time
		expected            assert.BoolAssertionFunc
	}{
		{
			desc:      "higher than 1/3 of the certificate lifetime left (lifetime > 10 days)",
			now:       time.Date(2025, 1, 19, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 30, 1, 1, 1, 1, time.UTC),
			expected:  assert.False,
		},
		{
			desc:      "lower than 1/3 of the certificate lifetime left(lifetime > 10 days)",
			now:       time.Date(2025, 1, 21, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 30, 1, 1, 1, 1, time.UTC),
			expected:  assert.True,
		},
		{
			desc:      "higher than 1/2 of the certificate lifetime left (lifetime < 10 days)",
			now:       time.Date(2025, 1, 4, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 10, 1, 1, 1, 1, time.UTC),
			expected:  assert.False,
		},
		{
			desc:      "lower than 1/2 of the certificate lifetime left (lifetime < 10 days)",
			now:       time.Date(2025, 1, 6, 1, 1, 1, 1, time.UTC),
			notBefore: time.Date(2025, 1, 1, 1, 1, 1, 1, time.UTC),
			notAfter:  time.Date(2025, 1, 10, 1, 1, 1, 1, time.UTC),
			expected:  assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			cert := &x509.Certificate{
				NotBefore: test.notBefore,
				NotAfter:  test.notAfter,
			}

			actual := isInRenewalPeriod(cert, "foo.com", noDays, test.now)

			test.expected(t, actual)
		})
	}
}

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
