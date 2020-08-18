package joker

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvUsername, EnvPassword).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success API key",
			envVars: map[string]string{
				EnvAPIKey: "123",
			},
		},
		{
			desc: "success username password",
			envVars: map[string]string{
				EnvUsername: "123",
				EnvPassword: "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey:   "",
				EnvUsername: "",
				EnvPassword: "",
			},
			expected: "joker: some credentials information are missing: JOKER_USERNAME,JOKER_PASSWORD or some credentials information are missing: JOKER_API_KEY",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvAPIKey:   "",
				EnvUsername: "123",
				EnvPassword: "",
			},
			expected: "joker: some credentials information are missing: JOKER_PASSWORD or some credentials information are missing: JOKER_API_KEY",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAPIKey:   "",
				EnvUsername: "",
				EnvPassword: "123",
			},
			expected: "joker: some credentials information are missing: JOKER_USERNAME or some credentials information are missing: JOKER_API_KEY",
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
				assert.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc            string
		apiKey          string
		username        string
		password        string
		baseURL         string
		expected        string
		expectedBaseURL string
	}{
		{
			desc:            "success api key",
			apiKey:          "123",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "success username and password",
			username:        "123",
			password:        "123",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "missing credentials",
			expected:        "joker: credentials missing",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "missing credentials: username",
			expected:        "joker: credentials missing",
			username:        "123",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "missing credentials: password",
			expected:        "joker: credentials missing",
			password:        "123",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "Base URL should ends with /",
			apiKey:          "123",
			baseURL:         "http://example.com",
			expectedBaseURL: "http://example.com/",
		},
		{
			desc:            "Base URL already ends with /",
			apiKey:          "123",
			baseURL:         "http://example.com/",
			expectedBaseURL: "http://example.com/",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.APIKey = test.apiKey
			config.Username = test.username
			config.Password = test.password

			if test.baseURL != "" {
				config.BaseURL = test.baseURL
			}

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.Equal(t, test.expectedBaseURL, p.config.BaseURL)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
