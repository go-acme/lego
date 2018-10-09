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
	dnsimpleLiveTest   bool
	dnsimpleOauthToken string
	dnsimpleDomain     string
	dnsimpleBaseURL    string
)

func init() {
	dnsimpleOauthToken = os.Getenv("DNSIMPLE_OAUTH_TOKEN")
	dnsimpleDomain = os.Getenv("DNSIMPLE_DOMAIN")
	dnsimpleBaseURL = "https://api.sandbox.dnsimple.com"

	if len(dnsimpleOauthToken) > 0 && len(dnsimpleDomain) > 0 {
		baseURL := os.Getenv("DNSIMPLE_BASE_URL")

		if baseURL != "" {
			dnsimpleBaseURL = baseURL
		}

		dnsimpleLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("DNSIMPLE_OAUTH_TOKEN", dnsimpleOauthToken)
	os.Setenv("DNSIMPLE_BASE_URL", dnsimpleBaseURL)
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

//
// Present
//

func TestLiveDNSimplePresent(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.AccessToken = dnsimpleOauthToken
	config.BaseURL = dnsimpleBaseURL

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(dnsimpleDomain, "", "123d==")
	assert.NoError(t, err)
}

//
// Cleanup
//

func TestLiveDNSimpleCleanUp(t *testing.T) {
	if !dnsimpleLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.AccessToken = dnsimpleOauthToken
	config.BaseURL = dnsimpleBaseURL

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.CleanUp(dnsimpleDomain, "", "123d==")
	assert.NoError(t, err)
}
