package dns01

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToFqdn(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		expected string
	}{
		{
			desc:     "simple",
			domain:   "foo.example.com",
			expected: "foo.example.com.",
		},
		{
			desc:     "already FQDN",
			domain:   "foo.example.com.",
			expected: "foo.example.com.",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fqdn := ToFqdn(test.domain)
			assert.Equal(t, test.expected, fqdn)
		})
	}
}

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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domain := UnFqdn(test.fqdn)

			assert.Equal(t, test.expected, domain)
		})
	}
}
