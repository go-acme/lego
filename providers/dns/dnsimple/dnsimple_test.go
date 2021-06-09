package dnsimple

import (
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sandboxURL = "https://api.sandbox.fake.com"

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvOAuthToken,
	EnvBaseURL).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvOAuthToken, envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvOAuthToken: "my_token",
			},
		},
		{
			desc: "success: base url",
			envVars: map[string]string{
				EnvOAuthToken: "my_token",
				EnvBaseURL:    "https://api.dnsimple.test",
			},
		},
		{
			desc: "missing oauth token",
			envVars: map[string]string{
				EnvOAuthToken: "",
			},
			expected: "dnsimple: OAuth token is missing",
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

				baseURL := os.Getenv(EnvBaseURL)
				if baseURL != "" {
					assert.Equal(t, baseURL, p.client.BaseURL)
				}
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		accessToken string
		baseURL     string
		expected    string
	}{
		{
			desc:        "success",
			accessToken: "my_token",
			baseURL:     "",
		},
		{
			desc:        "success: base url",
			accessToken: "my_token",
			baseURL:     "https://api.dnsimple.test",
		},
		{
			desc:     "missing oauth token",
			expected: "dnsimple: OAuth token is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessToken = test.accessToken
			config.BaseURL = test.baseURL

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)

				if test.baseURL != "" {
					assert.Equal(t, test.baseURL, p.client.BaseURL)
				}
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

	if os.Getenv(EnvBaseURL) == "" {
		os.Setenv(EnvBaseURL, sandboxURL)
	}

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

	if os.Getenv(EnvBaseURL) == "" {
		os.Setenv(EnvBaseURL, sandboxURL)
	}

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
