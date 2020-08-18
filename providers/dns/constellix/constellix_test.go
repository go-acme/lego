package constellix

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIKey,
	EnvSecretKey).
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
				EnvAPIKey:    "123",
				EnvSecretKey: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvSecretKey: "",
			},
			expected: "constellix: some credentials information are missing: CONSTELLIX_API_KEY,CONSTELLIX_SECRET_KEY",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvSecretKey: "api_password",
			},
			expected: "constellix: some credentials information are missing: CONSTELLIX_API_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAPIKey:    "api_username",
				EnvSecretKey: "",
			},
			expected: "constellix: some credentials information are missing: CONSTELLIX_SECRET_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider(nil)

			if len(test.expected) == 0 {
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
			desc:     "missing credentials",
			expected: "constellix: incomplete credentials, missing secret key and/or API key",
		},
		{
			desc:      "missing api key",
			apiKey:    "",
			secretKey: "api_secret",
			expected:  "constellix: incomplete credentials, missing secret key and/or API key",
		},
		{
			desc:      "missing secret key",
			apiKey:    "api_key",
			secretKey: "",
			expected:  "constellix: incomplete credentials, missing secret key and/or API key",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.APIKey = test.apiKey
			config.SecretKey = test.secretKey

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
