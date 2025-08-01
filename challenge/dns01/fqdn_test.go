package dns01

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnFqdn(t *testing.T) {
	testCases := []struct {
		desc     string
		fqdn     string
		expected string
	}{
		{
			desc:     "simple",
			fqdn:     "foo.example.",
			expected: "foo.example",
		},
		{
			desc:     "already domain",
			fqdn:     "foo.example",
			expected: "foo.example",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domain := UnFqdn(test.fqdn)

			assert.Equal(t, test.expected, domain)
		})
	}
}

func TestUnFqdnDomainsSeq(t *testing.T) {
	testCases := []struct {
		desc     string
		fqdn     string
		expected []string
	}{
		{
			desc:     "empty",
			fqdn:     "",
			expected: nil,
		},
		{
			desc:     "TLD",
			fqdn:     "com",
			expected: []string{"com"},
		},
		{
			desc:     "2 levels",
			fqdn:     "example.com",
			expected: []string{"example.com", "com"},
		},
		{
			desc:     "3 levels",
			fqdn:     "foo.example.com",
			expected: []string{"foo.example.com", "example.com", "com"},
		},
	}

	for _, test := range testCases {
		for name, suffix := range map[string]string{"": "", " FQDN": "."} { //nolint:gocritic
			t.Run(test.desc+name, func(t *testing.T) {
				t.Parallel()

				actual := slices.Collect(UnFqdnDomainsSeq(test.fqdn + suffix))

				assert.Equal(t, test.expected, actual)
			})
		}
	}
}

func TestDomainsSeq(t *testing.T) {
	testCases := []struct {
		desc     string
		fqdn     string
		expected []string
	}{
		{
			desc:     "empty",
			fqdn:     "",
			expected: nil,
		},
		{
			desc:     "empty FQDN",
			fqdn:     ".",
			expected: nil,
		},
		{
			desc:     "TLD FQDN",
			fqdn:     "com",
			expected: []string{"com"},
		},
		{
			desc:     "TLD",
			fqdn:     "com.",
			expected: []string{"com."},
		},
		{
			desc:     "2 levels",
			fqdn:     "example.com",
			expected: []string{"example.com", "com"},
		},
		{
			desc:     "2 levels FQDN",
			fqdn:     "example.com.",
			expected: []string{"example.com.", "com."},
		},
		{
			desc:     "3 levels",
			fqdn:     "foo.example.com",
			expected: []string{"foo.example.com", "example.com", "com"},
		},
		{
			desc:     "3 levels FQDN",
			fqdn:     "foo.example.com.",
			expected: []string{"foo.example.com.", "example.com.", "com."},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := slices.Collect(DomainsSeq(test.fqdn))

			assert.Equal(t, test.expected, actual)
		})
	}
}
