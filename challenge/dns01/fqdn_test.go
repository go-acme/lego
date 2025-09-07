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

	// Add tests for newer gTLD handling
	testCases = append(testCases, getNewerGTLDTestCases()...)

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := slices.Collect(DomainsSeq(test.fqdn))

			assert.Equal(t, test.expected, actual)
		})
	}
}

func getNewerGTLDTestCases() []struct {
	desc     string
	fqdn     string
	expected []string
} {
	return []struct {
		desc     string
		fqdn     string
		expected []string
	}{
		{
			desc:     ".dog gTLD - simple domain",
			fqdn:     "app4.dog",
			expected: []string{"app4.dog"},
		},
		{
			desc:     ".dog gTLD - simple domain FQDN",
			fqdn:     "app4.dog.",
			expected: []string{"app4.dog."},
		},
		{
			desc:     ".dog gTLD - subdomain",
			fqdn:     "play.app4.dog",
			expected: []string{"play.app4.dog", "app4.dog"},
		},
		{
			desc:     ".dog gTLD - deep subdomain",
			fqdn:     "_acme-challenge.play.app4.dog",
			expected: []string{"_acme-challenge.play.app4.dog", "play.app4.dog", "app4.dog"},
		},
		{
			desc:     ".app gTLD - subdomain", 
			fqdn:     "test.myapp.app",
			expected: []string{"test.myapp.app", "myapp.app"},
		},
		{
			desc:     "traditional TLD - unchanged behavior",
			fqdn:     "sub.example.com",
			expected: []string{"sub.example.com", "example.com", "com"},
		},
	}
}
