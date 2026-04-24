package rfc2136

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_altEnvNames(t *testing.T) {
	testCases := []struct {
		desc     string
		name     string
		expected []string
	}{
		{
			desc:     "Basic",
			name:     "DNSUPDATE_FOO",
			expected: []string{"RFC2136_FOO"},
		},
		{
			desc: "TSIG",
			name: "DNSUPDATE_TSIG_FOO",
			expected: []string{
				"RFC2136_TSIG_FOO",
			},
		},
		{
			desc: "TSIG_GSS",
			name: "DNSUPDATE_TSIG_GSS_FOO",
			expected: []string{
				"DNSUPDATE_RFC3645_FOO",
				"RFC2136_TSIG_GSS_FOO",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			names := altEnvNames(test.name)

			assert.Equal(t, test.expected, names)
		})
	}
}

func Test_getEnvStringSlice(t *testing.T) {
	t.Setenv("LEGO_TEST_EMPTY", "")
	t.Setenv("LEGO_TEST_SIMPLE", "bar")
	t.Setenv("LEGO_TEST_MULTIPLE", "foo,bar")

	testCases := []struct {
		desc     string
		name     string
		expected []string
	}{
		{
			desc: "non-existent env var",
			name: "LEGO_TEST_FOO",
		},
		{
			desc: "empty env var",
			name: "LEGO_TEST_EMPTY",
		},
		{
			desc:     "single value",
			name:     "LEGO_TEST_SIMPLE",
			expected: []string{"bar"},
		},
		{
			desc:     "multiple values",
			name:     "LEGO_TEST_MULTIPLE",
			expected: []string{"foo", "bar"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			v := getEnvStringSlice(test.name)

			assert.Equal(t, test.expected, v)
		})
	}
}
