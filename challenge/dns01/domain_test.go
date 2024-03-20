package dns01

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractSubDomain(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		zone     string
		expected string
	}{
		{
			desc:     "no FQDN",
			domain:   "_acme-challenge.example.com",
			zone:     "example.com",
			expected: "_acme-challenge",
		},
		{
			desc:     "no FQDN zone",
			domain:   "_acme-challenge.example.com.",
			zone:     "example.com",
			expected: "_acme-challenge",
		},
		{
			desc:     "no FQDN domain",
			domain:   "_acme-challenge.example.com",
			zone:     "example.com.",
			expected: "_acme-challenge",
		},
		{
			desc:     "FQDN",
			domain:   "_acme-challenge.example.com.",
			zone:     "example.com.",
			expected: "_acme-challenge",
		},
		{
			desc:     "multi-level subdomain",
			domain:   "_acme-challenge.one.example.com.",
			zone:     "example.com.",
			expected: "_acme-challenge.one",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			subDomain, err := ExtractSubDomain(test.domain, test.zone)
			require.NoError(t, err)

			assert.Equal(t, test.expected, subDomain)
		})
	}
}

func TestExtractSubDomain_errors(t *testing.T) {
	testCases := []struct {
		desc   string
		domain string
		zone   string
	}{
		{
			desc:   "same domain",
			domain: "example.com",
			zone:   "example.com",
		},
		{
			desc:   "same domain, no FQDN zone",
			domain: "example.com.",
			zone:   "example.com",
		},
		{
			desc:   "same domain, no FQDN domain",
			domain: "example.com",
			zone:   "example.com.",
		},
		{
			desc:   "same domain, FQDN",
			domain: "example.com.",
			zone:   "example.com.",
		},
		{
			desc:   "zone and domain are unrelated",
			domain: "_acme-challenge.example.com",
			zone:   "example.org",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := ExtractSubDomain(test.domain, test.zone)
			require.Error(t, err)
		})
	}
}
