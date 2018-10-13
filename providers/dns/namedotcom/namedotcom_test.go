package namedotcom

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	liveTest        bool
	envTestUsername string
	envTestAPIToken string
	envTestDomain   string
)

func init() {
	envTestUsername = os.Getenv("NAMECOM_USERNAME")
	envTestAPIToken = os.Getenv("NAMECOM_API_TOKEN")
	envTestDomain = os.Getenv("NAMEDOTCOM_DOMAIN")

	if len(envTestAPIToken) > 0 && len(envTestUsername) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("NAMECOM_USERNAME", envTestUsername)
	os.Setenv("NAMECOM_API_TOKEN", envTestAPIToken)
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
				"NAMECOM_USERNAME":  "A",
				"NAMECOM_API_TOKEN": "B",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"NAMECOM_USERNAME":  "",
				"NAMECOM_API_TOKEN": "",
			},
			expected: "namedotcom: some credentials information are missing: NAMECOM_USERNAME,NAMECOM_API_TOKEN",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				"NAMECOM_USERNAME":  "",
				"NAMECOM_API_TOKEN": "B",
			},
			expected: "namedotcom: some credentials information are missing: NAMECOM_USERNAME",
		},
		{
			desc: "missing api token",
			envVars: map[string]string{
				"NAMECOM_USERNAME":  "A",
				"NAMECOM_API_TOKEN": "",
			},
			expected: "namedotcom: some credentials information are missing: NAMECOM_API_TOKEN",
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
		apiToken string
		username string
		expected string
	}{
		{
			desc:     "success",
			apiToken: "A",
			username: "B",
		},
		{
			desc:     "missing credentials",
			expected: "namedotcom: username is required",
		},
		{
			desc:     "missing API token",
			apiToken: "",
			username: "B",
			expected: "namedotcom: API token is required",
		},
		{
			desc:     "missing username",
			apiToken: "A",
			username: "",
			expected: "namedotcom: username is required",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("NAMECOM_USERNAME")
			os.Unsetenv("NAMECOM_API_TOKEN")

			config := NewDefaultConfig()
			config.Username = test.username
			config.APIToken = test.apiToken

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
