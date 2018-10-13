package sakuracloud

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest            bool
	envTestAccessToken  string
	envTestAccessSecret string
	envTestDomain       string
)

func init() {
	envTestAccessToken = os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	envTestAccessSecret = os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
	envTestDomain = os.Getenv("SAKURACLOUD_DOMAIN")

	if len(envTestAccessToken) > 0 && len(envTestAccessSecret) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN", envTestAccessToken)
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", envTestAccessSecret)
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
				"SAKURACLOUD_ACCESS_TOKEN":        "123",
				"SAKURACLOUD_ACCESS_TOKEN_SECRET": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"SAKURACLOUD_ACCESS_TOKEN":        "",
				"SAKURACLOUD_ACCESS_TOKEN_SECRET": "",
			},
			expected: "sakuracloud: some credentials information are missing: SAKURACLOUD_ACCESS_TOKEN,SAKURACLOUD_ACCESS_TOKEN_SECRET",
		},
		{
			desc: "missing access token",
			envVars: map[string]string{
				"SAKURACLOUD_ACCESS_TOKEN":        "",
				"SAKURACLOUD_ACCESS_TOKEN_SECRET": "456",
			},
			expected: "sakuracloud: some credentials information are missing: SAKURACLOUD_ACCESS_TOKEN",
		},
		{
			desc: "missing token secret",
			envVars: map[string]string{
				"SAKURACLOUD_ACCESS_TOKEN":        "123",
				"SAKURACLOUD_ACCESS_TOKEN_SECRET": "",
			},
			expected: "sakuracloud: some credentials information are missing: SAKURACLOUD_ACCESS_TOKEN_SECRET",
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
		desc     string
		token    string
		secret   string
		expected string
	}{
		{
			desc:   "success",
			token:  "123",
			secret: "456",
		},
		{
			desc:     "missing credentials",
			expected: "sakuracloud: AccessToken is missing",
		},
		{
			desc:     "missing token",
			secret:   "456",
			expected: "sakuracloud: AccessToken is missing",
		},
		{
			desc:     "missing secret",
			token:    "123",
			expected: "sakuracloud: AccessSecret is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN")
			os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

			config := NewDefaultConfig()
			config.Token = test.token
			config.Secret = test.secret

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
