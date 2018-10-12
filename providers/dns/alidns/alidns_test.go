package alidns

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
	envTestAPIKey = os.Getenv("ALICLOUD_ACCESS_KEY")
	envTestSecretKey = os.Getenv("ALICLOUD_SECRET_KEY")
	envTestDomain = os.Getenv("ALIDNS_DOMAIN")

	if len(envTestAPIKey) > 0 && len(envTestSecretKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("ALICLOUD_ACCESS_KEY", envTestAPIKey)
	os.Setenv("ALICLOUD_SECRET_KEY", envTestSecretKey)
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
				"ALICLOUD_ACCESS_KEY": "123",
				"ALICLOUD_SECRET_KEY": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"ALICLOUD_ACCESS_KEY": "",
				"ALICLOUD_SECRET_KEY": "",
			},
			expected: "alicloud: some credentials information are missing: ALICLOUD_ACCESS_KEY,ALICLOUD_SECRET_KEY",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				"ALICLOUD_ACCESS_KEY": "",
				"ALICLOUD_SECRET_KEY": "456",
			},
			expected: "alicloud: some credentials information are missing: ALICLOUD_ACCESS_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				"ALICLOUD_ACCESS_KEY": "123",
				"ALICLOUD_SECRET_KEY": "",
			},
			expected: "alicloud: some credentials information are missing: ALICLOUD_SECRET_KEY",
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
			expected: "alicloud: credentials missing",
		},
		{
			desc:      "missing api key",
			secretKey: "456",
			expected:  "alicloud: credentials missing",
		},
		{
			desc:     "missing secret key",
			apiKey:   "123",
			expected: "alicloud: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("ALICLOUD_ACCESS_KEY")
			os.Unsetenv("ALICLOUD_SECRET_KEY")

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
