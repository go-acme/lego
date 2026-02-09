package combell

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvAPISecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey:    "key",
				EnvAPISecret: "secret",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPISecret: "secret",
			},
			expected: "combell: some credentials information are missing: COMBELL_API_KEY",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvAPIKey: "key",
			},
			expected: "combell: some credentials information are missing: COMBELL_API_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "combell: some credentials information are missing: COMBELL_API_KEY,COMBELL_API_SECRET",
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
		desc      string
		apiKey    string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "key",
			apiSecret: "secret",
		},
		{
			desc:      "missing API key",
			apiSecret: "secret",
			expected:  "combell: some credentials information are missing",
		},
		{
			desc:     "missing API secret",
			apiKey:   "key",
			expected: "combell: some credentials information are missing",
		},
		{
			desc:     "missing credentials",
			expected: "combell: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.APISecret = test.apiSecret

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
