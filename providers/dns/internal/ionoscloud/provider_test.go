package ionoscloud

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		token    string
		tll      int
		expected string
	}{
		{
			desc:  "success",
			token: "123",
			tll:   MinTTL,
		},
		{
			desc:     "missing credentials",
			tll:      MinTTL,
			expected: "credentials missing",
		},
		{
			desc:     "invalid TTL",
			token:    "123",
			tll:      30,
			expected: "invalid TTL, TTL (30) must be greater than 300",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.Token = test.token
			config.TTL = test.tll

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
