package pdns

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	liveTest      bool
	envTestAPIURL *url.URL
	envTestAPIKey string
	envTestDomain string
)

func init() {
	envTestAPIURL, _ = url.Parse(os.Getenv("PDNS_API_URL"))
	envTestAPIKey = os.Getenv("PDNS_API_KEY")
	envTestDomain = os.Getenv("PDNS_DOMAIN")

	if len(envTestAPIURL.String()) > 0 && len(envTestAPIKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("PDNS_API_URL", envTestAPIURL.String())
	os.Setenv("PDNS_API_KEY", envTestAPIKey)
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
				"PDNS_API_KEY": "123",
				"PDNS_API_URL": "http://example.com",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"PDNS_API_KEY": "",
				"PDNS_API_URL": "",
			},
			expected: "pdns: some credentials information are missing: PDNS_API_KEY,PDNS_API_URL",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				"PDNS_API_KEY": "",
				"PDNS_API_URL": "http://example.com",
			},
			expected: "pdns: some credentials information are missing: PDNS_API_KEY",
		},
		{
			desc: "missing API URL",
			envVars: map[string]string{
				"PDNS_API_KEY": "123",
				"PDNS_API_URL": "",
			},
			expected: "pdns: some credentials information are missing: PDNS_API_URL",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		host     *url.URL
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
			host: func() *url.URL {
				u, _ := url.Parse("http://example.com")
				return u
			}(),
		},
		{
			desc:     "missing credentials",
			expected: "pdns: API key missing",
		},
		{
			desc:   "missing API key",
			apiKey: "",
			host: func() *url.URL {
				u, _ := url.Parse("http://example.com")
				return u
			}(),
			expected: "pdns: API key missing",
		},
		{
			desc:     "missing host",
			apiKey:   "123",
			expected: "pdns: API URL missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("PDNS_API_KEY")
			os.Unsetenv("PDNS_API_URL")

			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.Host = test.host

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
