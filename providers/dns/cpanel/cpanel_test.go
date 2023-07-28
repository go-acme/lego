package cpanel

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvToken,
	EnvBaseURL,
	EnvNameserver).
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
				EnvToken:      "secret",
				EnvBaseURL:    "https://example.com",
				EnvNameserver: "ns.example.com:53",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvBaseURL:    "https://example.com",
				EnvNameserver: "ns.example.com:53",
			},
			expected: "cpanel: some credentials information are missing: CPANEL_TOKEN",
		},
		{
			desc: "missing base URL",
			envVars: map[string]string{
				EnvToken:      "secret",
				EnvBaseURL:    "",
				EnvNameserver: "ns.example.com:53",
			},
			expected: "cpanel: some credentials information are missing: CPANEL_BASE_URL",
		},
		{
			desc: "missing nameserver",
			envVars: map[string]string{
				EnvToken:      "secret",
				EnvBaseURL:    "https://example.com",
				EnvNameserver: "",
			},
			expected: "cpanel: some credentials information are missing: CPANEL_NAMESERVER",
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
		desc       string
		username   string
		token      string
		baseURL    string
		nameserver string
		expected   string
	}{
		{
			desc:       "success",
			username:   "user",
			token:      "secret",
			baseURL:    "https://example.com",
			nameserver: "ns.example.com:53",
		},
		{
			desc:       "missing username",
			username:   "",
			token:      "secret",
			baseURL:    "https://example.com",
			nameserver: "ns.example.com:53",
			expected:   "cpanel: some credentials information are missing",
		},
		{
			desc:       "missing token",
			username:   "user",
			token:      "",
			baseURL:    "https://example.com",
			nameserver: "ns.example.com:53",
			expected:   "cpanel: some credentials information are missing",
		},
		{
			desc:       "missing base URL",
			username:   "user",
			token:      "secret",
			baseURL:    "",
			nameserver: "ns.example.com:53",
			expected:   "cpanel: server information are missing",
		},
		{
			desc:       "missing nameserver",
			username:   "user",
			token:      "secret",
			baseURL:    "https://example.com",
			nameserver: "",
			expected:   "cpanel: server information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Token = test.token
			config.BaseURL = test.baseURL
			config.Nameserver = test.nameserver

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
