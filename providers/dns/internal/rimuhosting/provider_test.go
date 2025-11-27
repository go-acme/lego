package rimuhosting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		expected  string
		apiKey    string
		secretKey string
	}{
		{
			desc:      "success",
			apiKey:    "api_key",
			secretKey: "api_secret",
		},
		{
			desc:      "missing api key",
			apiKey:    "",
			secretKey: "api_secret",
			expected:  "incomplete credentials, missing API key",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config, "")

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}
