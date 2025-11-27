package selectel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		token    string
		ttl      int
		expected string
	}{
		{
			desc:  "success",
			token: "123",
			ttl:   60,
		},
		{
			desc:     "missing api key",
			token:    "",
			ttl:      60,
			expected: "credentials missing",
		},
		{
			desc:     "bad TTL value",
			token:    "123",
			ttl:      59,
			expected: fmt.Sprintf("invalid TTL, TTL (59) must be greater than %d", MinTTL),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.TTL = test.ttl
			config.Token = test.token

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}
