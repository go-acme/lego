package technitium

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvServerBaseURL, EnvAPIToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvServerBaseURL: "https://localhost:5380",
				EnvAPIToken:      "secret",
			},
		},
		{
			desc: "missing server base URL",
			envVars: map[string]string{
				EnvServerBaseURL: "",
				EnvAPIToken:      "secret",
			},
			expected: "technitium: some credentials information are missing: TECHNITIUM_SERVER_BASE_URL",
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvServerBaseURL: "https://localhost:5380",
				EnvAPIToken:      "",
			},
			expected: "technitium: some credentials information are missing: TECHNITIUM_API_TOKEN",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "technitium: some credentials information are missing: TECHNITIUM_SERVER_BASE_URL,TECHNITIUM_API_TOKEN",
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
	testCases := []struct {
		desc     string
		baseURL  string
		token    string
		expected string
	}{
		{
			desc:    "success",
			baseURL: "https://localhost:5380",
			token:   "secret",
		},
		{
			desc:     "missing server base URL",
			token:    "secret",
			expected: "technitium: missing server URL",
		},
		{
			desc:     "missing token",
			baseURL:  "https://localhost:5380",
			expected: "technitium: missing credentials",
		},
		{
			desc:     "missing credentials",
			expected: "technitium: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = test.baseURL
			config.APIToken = test.token

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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
