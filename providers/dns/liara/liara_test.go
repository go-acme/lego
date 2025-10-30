package liara

import (
	"fmt"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const (
	envDomain         = envNamespace + "DOMAIN"
	lowerThanMinTTL   = 100
	greaterThanMaxTTL = 440000
)

var envTest = tester.NewEnvTest(EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "key",
			},
		},
		{
			desc:     "missing API key",
			envVars:  map[string]string{},
			expected: "liara: some credentials information are missing: LIARA_API_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

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

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		ttl      int
		expected string
	}{
		{
			desc:   "success",
			apiKey: "key",
			ttl:    minTTL,
		},
		{
			desc:     "missing API key",
			ttl:      maxTTL,
			expected: "liara: APIKey is missing",
		},
		{
			desc:     "invalid TTL",
			ttl:      lowerThanMinTTL,
			apiKey:   "key",
			expected: fmt.Sprintf("liara: invalid TTL, TTL (%d) must be greater than %d", lowerThanMinTTL, minTTL),
		},
		{
			desc:     "invalid TTL",
			ttl:      greaterThanMaxTTL,
			apiKey:   "key",
			expected: fmt.Sprintf("liara: invalid TTL, TTL (%d) must be lower than %d", greaterThanMaxTTL, maxTTL),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.TTL = test.ttl

			p, err := NewDNSProviderConfig(config)

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

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
