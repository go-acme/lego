package ispconfig

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvEndpoint,
	EnvToken).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvEndpoint: "https://example.com",
				EnvToken:    "secret",
			},
		},
		{
			desc: "missing endpoint",
			envVars: map[string]string{
				EnvEndpoint: "",
				EnvToken:    "secret",
			},
			expected: "ispconfig: some credentials information are missing: ISPCONFIG_ENDPOINT",
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvEndpoint: "https://example.com",
				EnvToken:    "",
			},
			expected: "ispconfig: some credentials information are missing: ISPCONFIG_TOKEN",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "ispconfig: some credentials information are missing: ISPCONFIG_ENDPOINT,ISPCONFIG_TOKEN",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		endpoint string
		token    string
		expected string
	}{
		{
			desc:     "success",
			endpoint: "https://example.com",
			token:    "secret",
		},
		{
			desc:     "missing endpoint",
			endpoint: "",
			token:    "secret",
			expected: "ispconfig: missing endpoint",
		},
		{
			desc:     "missing token",
			endpoint: "https://example.com",
			token:    "",
			expected: "ispconfig: missing token",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Endpoint = test.endpoint
			config.Token = test.token

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}
