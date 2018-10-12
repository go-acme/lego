package cloudxns

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest         bool
	envTestAPIKey    string
	envTestSecretKey string
	envTestDomain    string
)

func init() {
	envTestAPIKey = os.Getenv("CLOUDXNS_API_KEY")
	envTestSecretKey = os.Getenv("CLOUDXNS_SECRET_KEY")
	envTestDomain = os.Getenv("CLOUDXNS_DOMAIN")

	if len(envTestAPIKey) > 0 && len(envTestSecretKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("CLOUDXNS_API_KEY", envTestAPIKey)
	os.Setenv("CLOUDXNS_SECRET_KEY", envTestSecretKey)
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
				"CLOUDXNS_API_KEY":    "123",
				"CLOUDXNS_SECRET_KEY": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"CLOUDXNS_API_KEY":    "",
				"CLOUDXNS_SECRET_KEY": "",
			},
			expected: "CloudXNS: some credentials information are missing: CLOUDXNS_API_KEY,CLOUDXNS_SECRET_KEY",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				"CLOUDXNS_API_KEY":    "",
				"CLOUDXNS_SECRET_KEY": "456",
			},
			expected: "CloudXNS: some credentials information are missing: CLOUDXNS_API_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				"CLOUDXNS_API_KEY":    "123",
				"CLOUDXNS_SECRET_KEY": "",
			},
			expected: "CloudXNS: some credentials information are missing: CLOUDXNS_SECRET_KEY",
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
		apiKey    string
		secretKey string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "123",
			secretKey: "456",
		},
		{
			desc:     "missing credentials",
			expected: "CloudXNS: credentials missing: apiKey",
		},
		{
			desc:      "missing api key",
			secretKey: "456",
			expected:  "CloudXNS: credentials missing: apiKey",
		},
		{
			desc:     "missing secret key",
			apiKey:   "123",
			expected: "CloudXNS: credentials missing: secretKey",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("CLOUDXNS_API_KEY")
			os.Unsetenv("CLOUDXNS_SECRET_KEY")

			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.SecretKey = test.secretKey

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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

func TestPresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(envTestAPIKey, envTestSecretKey)
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 2)

	provider, err := NewDNSProviderCredentials(envTestAPIKey, envTestSecretKey)
	require.NoError(t, err)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
