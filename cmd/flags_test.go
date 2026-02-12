package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_toEnvName(t *testing.T) {
	testCases := []struct {
		desc     string
		flag     string
		expected string
	}{
		{
			desc:     "only letters",
			flag:     flgServer,
			expected: "LEGO_SERVER",
		},
		{
			desc:     "letters and digits",
			flag:     flgIPv6Only,
			expected: "LEGO_IPV6ONLY",
		},
		{
			desc:     "hyphen",
			flag:     flgHTTPPort,
			expected: "LEGO_HTTP_PORT",
		},
		{
			desc:     "dot, hyphen",
			flag:     flgDNSPropagationDisableRNS,
			expected: "LEGO_DNS_PROPAGATION_DISABLE_RNS",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			envName := toEnvName(test.flag)

			assert.Equal(t, test.expected, envName)
		})
	}
}
