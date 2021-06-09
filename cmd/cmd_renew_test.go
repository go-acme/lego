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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := merge(test.prevDomains, test.nextDomains)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_needRenewal(t *testing.T) {
	testCases := []struct {
		desc     string
		x509Cert *x509.Certificate
		days     int
		expected bool
	}{
		{
			desc: "30 days, NotAfter now",
			x509Cert: &x509.Certificate{
				NotAfter: time.Now(),
			},
			days:     30,
			expected: true,
		},
		{
			desc: "30 days, NotAfter 31 days",
			x509Cert: &x509.Certificate{
				NotAfter: time.Now().Add(31*24*time.Hour + 1*time.Second),
			},
			days:     30,
			expected: false,
		},
		{
			desc: "30 days, NotAfter 30 days",
			x509Cert: &x509.Certificate{
				NotAfter: time.Now().Add(30 * 24 * time.Hour),
			},
			days:     30,
			expected: true,
		},
		{
			desc: "0 days, NotAfter 30 days: only the day of the expiration",
			x509Cert: &x509.Certificate{
				NotAfter: time.Now().Add(30 * 24 * time.Hour),
			},
			days:     0,
			expected: false,
		},
		{
			desc: "-1 days, NotAfter 30 days: always renew",
			x509Cert: &x509.Certificate{
				NotAfter: time.Now().Add(30 * 24 * time.Hour),
			},
			days:     -1,
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			actual := needRenewal(test.x509Cert, "foo.com", test.days)

			assert.Equal(t, test.expected, actual)
		})
	}
}
