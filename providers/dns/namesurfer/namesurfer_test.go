package namesurfer

import (
	"testing"

	"github.com/go-acme/lego/v5/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvBaseURL,
	EnvAPIKey,
	EnvAPISecret,
	EnvView,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvBaseURL:   "https://example.com",
				EnvAPIKey:    "user",
				EnvAPISecret: "secret",
			},
		},
		{
			desc: "missing base URL",
			envVars: map[string]string{
				EnvBaseURL:   "",
				EnvAPIKey:    "user",
				EnvAPISecret: "secret",
			},
			expected: "namesurfer: some credentials information are missing: NAMESURFER_BASE_URL",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvBaseURL:   "https://example.com",
				EnvAPIKey:    "",
				EnvAPISecret: "secret",
			},
			expected: "namesurfer: some credentials information are missing: NAMESURFER_API_KEY",
		},
		{
			desc: "missing API secret",
			envVars: map[string]string{
				EnvBaseURL:   "https://example.com",
				EnvAPIKey:    "user",
				EnvAPISecret: "",
			},
			expected: "namesurfer: some credentials information are missing: NAMESURFER_API_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "namesurfer: some credentials information are missing: NAMESURFER_BASE_URL,NAMESURFER_API_KEY,NAMESURFER_API_SECRET",
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
		desc      string
		baseURL   string
		apiKey    string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			baseURL:   "https://example.com",
			apiKey:    "user",
			apiSecret: "secret",
		},
		{
			desc:      "missing base URL",
			apiKey:    "user",
			apiSecret: "secret",
			expected:  "namesurfer: base URL missing",
		},
		{
			desc:      "missing API key",
			baseURL:   "https://example.com",
			apiSecret: "secret",
			expected:  "namesurfer: credentials missing",
		},
		{
			desc:     "missing API secret",
			baseURL:  "https://example.com",
			apiKey:   "user",
			expected: "namesurfer: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "namesurfer: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = test.baseURL
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

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
