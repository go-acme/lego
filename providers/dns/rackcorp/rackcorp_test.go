package rackcorp

import (
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIURL, EnvAPIUUID, EnvAPISecret).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvAPIUUID, EnvAPISecret, envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUUID:   "my_uuid",
				EnvAPISecret: "my_secret",
			},
		},
		{
			desc: "success: url",
			envVars: map[string]string{
				EnvAPIUUID:   "my_uuid",
				EnvAPISecret: "my_secret",
				EnvAPIURL:    "https://api.rackcorp.net/api/rest/v2.8/json.php",
			},
		},
		{
			desc: "missing API credentials",
			envVars: map[string]string{
				EnvAPIUUID:   "",
				EnvAPISecret: "",
			},
			expected: "rackcorp: API credentials are missing",
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

				baseURL := os.Getenv(EnvAPIURL)
				if baseURL != "" {
					assert.Equal(t, baseURL, p.client.URL)
				}
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		apiUUID   string
		apiSecret string
		url       string
		expected  string
	}{
		{
			desc:      "success",
			apiUUID:   "my_uuid",
			apiSecret: "my_secret",
			url:       "",
		},
		{
			desc:      "success: url",
			apiUUID:   "my_uuid",
			apiSecret: "my_secret",
			url:       "https://api.rackcorp.net/api/rest/v2.8/json.php",
		},
		{
			desc:     "missing API credentials",
			expected: "rackcorp: API credentials are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUUID = test.apiUUID
			config.APISecret = test.apiSecret
			config.URL = test.url

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)

				if test.url != "" {
					assert.Equal(t, test.url, p.client.URL)
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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
