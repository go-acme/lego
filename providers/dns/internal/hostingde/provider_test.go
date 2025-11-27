package hostingde

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		zoneName string
		expected string
	}{
		{
			desc:     "success",
			apiKey:   "123",
			zoneName: "example.org",
		},
		{
			desc:     "missing credentials",
			expected: "API key missing",
		},
		{
			desc:     "missing api key",
			zoneName: "456",
			expected: "API key missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.APIKey = test.apiKey
			config.ZoneName = test.zoneName

			p, err := NewDNSProviderConfig(config, "")

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}
