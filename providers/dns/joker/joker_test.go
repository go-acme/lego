package joker

import (
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest("JOKER_API_KEY").WithDomain("JOKER_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"JOKER_API_KEY": "123",
			},
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				"JOKER_API_KEY": "",
			},
			expected: "joker: some credentials information are missing: JOKER_API_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

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
		authKey         string
		baseURL         string
		expected        string
		expectedBaseURL string
	}{
		{
			desc:            "success",
			authKey:         "123",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "missing credentials",
			expected:        "joker: credentials missing",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "Base URL should ends with /",
			authKey:         "123",
			baseURL:         "http://example.com",
			expectedBaseURL: "http://example.com/",
		},
		{
			desc:            "Base URL already ends with /",
			authKey:         "123",
			baseURL:         "http://example.com/",
			expectedBaseURL: "http://example.com/",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.authKey
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
