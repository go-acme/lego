package dnsupdate

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
