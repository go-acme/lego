package dnspod

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest      bool
	envTestAPIKey string
	envTestDomain string
)

func init() {
	envTestAPIKey = os.Getenv("DNSPOD_API_KEY")
	envTestDomain = os.Getenv("DNSPOD_DOMAIN")

	if len(envTestAPIKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("DNSPOD_API_KEY", envTestAPIKey)
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
				"DNSPOD_API_KEY": "123",
			},
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				"DNSPOD_API_KEY": "",
			},
			expected: "dnspod: some credentials information are missing: DNSPOD_API_KEY",
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
		desc       string
		loginToken string
		expected   string
	}{
		{
			desc:       "success",
			loginToken: "123",
		},
		{
			desc:     "missing credentials",
			expected: "dnspod: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("DNSPOD_API_KEY")

			config := NewDefaultConfig()
			config.LoginToken = test.loginToken

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
