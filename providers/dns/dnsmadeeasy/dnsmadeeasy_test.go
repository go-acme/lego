package dnsmadeeasy

import (
	"os"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIKey,
	EnvAPISecret).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	os.Setenv(EnvSandbox, "true")

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey:    "123",
				EnvAPISecret: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvAPISecret: "",
			},
			expected: "dnsmadeeasy: some credentials information are missing: DNSMADEEASY_API_KEY,DNSMADEEASY_API_SECRET",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvAPISecret: "456",
			},
			expected: "dnsmadeeasy: some credentials information are missing: DNSMADEEASY_API_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAPIKey:    "123",
				EnvAPISecret: "",
			},
			expected: "dnsmadeeasy: some credentials information are missing: DNSMADEEASY_API_SECRET",
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
	os.Setenv(EnvSandbox, "true")

	testCases := []struct {
		desc      string
		apiKey    string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "123",
			apiSecret: "456",
		},
		{
			desc:     "missing credentials",
			expected: "dnsmadeeasy: credentials missing: API key",
		},
		{
			desc:      "missing api key",
			apiSecret: "456",
			expected:  "dnsmadeeasy: credentials missing: API key",
		},
		{
			desc:     "missing secret key",
			apiKey:   "123",
			expected: "dnsmadeeasy: credentials missing: API secret",
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

func TestLivePresentAndCleanup(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	os.Setenv(EnvSandbox, "true")

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
