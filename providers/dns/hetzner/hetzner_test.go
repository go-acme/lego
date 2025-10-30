package hetzner

import (
	"testing"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/providers/dns/hetzner/internal/hetznerv1"
	"github.com/go-acme/lego/v4/providers/dns/hetzner/internal/legacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAPIKey, EnvAPIToken)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc    string
		envVars map[string]string

		expectedProvider challenge.ProviderTimeout
		expectedError    string
	}{
		{
			desc: "success (v1)",
			envVars: map[string]string{
				EnvAPIToken: "123",
			},
			expectedProvider: &hetznerv1.DNSProvider{},
		},
		{
			desc: "success (legacy)",
			envVars: map[string]string{
				EnvAPIKey: "123",
			},
			expectedProvider: &legacy.DNSProvider{},
		},
		{
			desc: "success (both)",
			envVars: map[string]string{
				EnvAPIKey:   "123",
				EnvAPIToken: "123",
			},
			expectedProvider: &hetznerv1.DNSProvider{},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey:   "",
				EnvAPIToken: "",
			},
			expectedError: "hetzner: some credentials information are missing: HETZNER_API_TOKEN",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expectedError == "" {
				require.NoError(t, err)
				assert.IsType(t, test.expectedProvider, p.provider)
				require.NotNil(t, p)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		apiToken string
		ttl      int

		expectedProvider challenge.ProviderTimeout
		expectedError    string
	}{
		{
			desc:             "success (v1)",
			ttl:              minTTL,
			apiToken:         "123",
			expectedProvider: &hetznerv1.DNSProvider{},
		},
		{
			desc:             "success (legacy)",
			ttl:              minTTL,
			apiKey:           "456",
			expectedProvider: &legacy.DNSProvider{},
		},
		{
			desc:             "success (both)",
			ttl:              minTTL,
			apiToken:         "123",
			apiKey:           "456",
			expectedProvider: &hetznerv1.DNSProvider{},
		},
		{
			desc:          "missing credentials",
			ttl:           minTTL,
			expectedError: "hetzner: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIToken = test.apiToken
			config.APIKey = test.apiKey
			config.TTL = test.ttl

			p, err := NewDNSProviderConfig(config)

			if test.expectedError == "" {
				require.NoError(t, err)
				assert.IsType(t, test.expectedProvider, p.provider)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}
