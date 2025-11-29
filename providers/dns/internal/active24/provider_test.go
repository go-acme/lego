package active24

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		secret   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "user",
			secret: "secret",
		},
		{
			desc:     "missing API key",
			apiKey:   "",
			secret:   "secret",
			expected: "credentials missing",
		},
		{
			desc:     "missing secret",
			apiKey:   "user",
			secret:   "",
			expected: "credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.APIKey = test.apiKey
			config.Secret = test.secret

			p, err := NewDNSProviderConfig(config, "example.com")

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
