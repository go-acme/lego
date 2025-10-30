package porkbun

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvSecretAPIKey, EnvAPIKey).
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
				EnvSecretAPIKey: "secret",
				EnvAPIKey:       "key",
			},
		},
		{
			desc: "missing secret API key",
			envVars: map[string]string{
				EnvSecretAPIKey: "",
				EnvAPIKey:       "key",
			},
			expected: "porkbun: some credentials information are missing: PORKBUN_SECRET_API_KEY",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvSecretAPIKey: "secret",
				EnvAPIKey:       "",
			},
			expected: "porkbun: some credentials information are missing: PORKBUN_API_KEY",
		},
		{
			desc: "missing all credentials",
			envVars: map[string]string{
				EnvSecretAPIKey: "",
				EnvAPIKey:       "",
			},
			expected: "porkbun: some credentials information are missing: PORKBUN_SECRET_API_KEY,PORKBUN_API_KEY",
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
		desc         string
		secretAPIKey string
		apiKey       string
		expected     string
	}{
		{
			desc:         "success",
			secretAPIKey: "secret",
			apiKey:       "key",
		},
		{
			desc:     "missing secret API key",
			apiKey:   "key",
			expected: "porkbun: some credentials information are missing",
		},
		{
			desc:         "missing API key",
			secretAPIKey: "secret",
			expected:     "porkbun: some credentials information are missing",
		},
		{
			desc:     "missing all credentials",
			expected: "porkbun: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.SecretAPIKey = test.secretAPIKey
			config.APIKey = test.apiKey

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
