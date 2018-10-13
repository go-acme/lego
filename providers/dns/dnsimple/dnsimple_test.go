package dnsimple

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/acme"
)

var (
	liveTest          bool
	envTestOauthToken string
	envTestDomain     string
	envTestBaseURL    string
)

func init() {
	envTestOauthToken = os.Getenv("DNSIMPLE_OAUTH_TOKEN")
	envTestDomain = os.Getenv("DNSIMPLE_DOMAIN")
	envTestBaseURL = "https://api.sandbox.fake.com"

	if len(envTestOauthToken) > 0 && len(envTestDomain) > 0 {
		baseURL := os.Getenv("DNSIMPLE_BASE_URL")

		if baseURL != "" {
			envTestBaseURL = baseURL
		}

		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("DNSIMPLE_OAUTH_TOKEN", envTestOauthToken)
	os.Setenv("DNSIMPLE_BASE_URL", envTestBaseURL)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc      string
		userAgent string
		envVars   map[string]string
		expected  string
	}{
		{
			desc:      "success",
			userAgent: "lego",
			envVars: map[string]string{
				"DNSIMPLE_OAUTH_TOKEN": "my_token",
			},
		},
		{
			desc: "success: base url",
			envVars: map[string]string{
				"DNSIMPLE_OAUTH_TOKEN": "my_token",
				"DNSIMPLE_BASE_URL":    "https://api.dnsimple.test",
			},
		},
		{
			desc: "missing oauth token",
			envVars: map[string]string{
				"DNSIMPLE_OAUTH_TOKEN": "",
			},
			expected: "dnsimple: OAuth token is missing",
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

			if test.userAgent != "" {
				acme.UserAgent = test.userAgent
			}

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)

				baseURL := os.Getenv("DNSIMPLE_BASE_URL")
				if baseURL != "" {
					assert.Equal(t, baseURL, p.client.BaseURL)
				}

				if test.userAgent != "" {
					assert.Equal(t, "lego", p.client.UserAgent)
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
			defer restoreEnv()
			os.Unsetenv("DNSIMPLE_OAUTH_TOKEN")
			os.Unsetenv("DNSIMPLE_BASE_URL")

			config := NewDefaultConfig()
			config.AccessToken = test.accessToken
			config.BaseURL = test.baseURL

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
