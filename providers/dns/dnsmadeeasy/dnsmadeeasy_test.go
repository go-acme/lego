package dnsmadeeasy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	liveTest         bool
	envTestAPIKey    string
	envTestAPISecret string
	envTestDomain    string
)

func init() {
	envTestAPIKey = os.Getenv("DNSMADEEASY_API_KEY")
	envTestAPISecret = os.Getenv("DNSMADEEASY_API_SECRET")
	envTestDomain = os.Getenv("DNSMADEEASY_DOMAIN")

	os.Setenv("DNSMADEEASY_SANDBOX", "true")

	liveTest = len(envTestAPIKey) > 0 && len(envTestAPISecret) > 0
}

func restoreEnv() {
	os.Setenv("DNSMADEEASY_API_KEY", envTestAPIKey)
	os.Setenv("DNSMADEEASY_API_SECRET", envTestAPISecret)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"DNSMADEEASY_API_KEY":    "123",
				"DNSMADEEASY_API_SECRET": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"DNSMADEEASY_API_KEY":    "",
				"DNSMADEEASY_API_SECRET": "",
			},
			expected: "dnsmadeeasy: some credentials information are missing: DNSMADEEASY_API_KEY,DNSMADEEASY_API_SECRET",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				"DNSMADEEASY_API_KEY":    "",
				"DNSMADEEASY_API_SECRET": "456",
			},
			expected: "dnsmadeeasy: some credentials information are missing: DNSMADEEASY_API_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				"DNSMADEEASY_API_KEY":    "123",
				"DNSMADEEASY_API_SECRET": "",
			},
			expected: "dnsmadeeasy: some credentials information are missing: DNSMADEEASY_API_SECRET",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			for key, value := range test.envVars {
				if len(value) == 0 {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
			defer restoreEnv()
			os.Unsetenv("DNSMADEEASY_API_KEY")
			os.Unsetenv("DNSMADEEASY_API_SECRET")

			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.APISecret = test.apiSecret

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
