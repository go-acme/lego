package glesys

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	liveTest       bool
	envTestAPIUser string
	envTestAPIKey  string
	envTestDomain  string
)

func init() {
	envTestAPIUser = os.Getenv("GLESYS_API_USER")
	envTestAPIKey = os.Getenv("GLESYS_API_KEY")

	if len(envTestAPIUser) > 0 && len(envTestAPIKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("GLESYS_API_USER", envTestAPIUser)
	os.Setenv("GLESYS_API_KEY", envTestAPIKey)
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
				"GLESYS_API_USER": "A",
				"GLESYS_API_KEY":  "B",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"GLESYS_API_USER": "",
				"GLESYS_API_KEY":  "",
			},
			expected: "glesys: some credentials information are missing: GLESYS_API_USER,GLESYS_API_KEY",
		},
		{
			desc: "missing api user",
			envVars: map[string]string{
				"GLESYS_API_USER": "",
				"GLESYS_API_KEY":  "B",
			},
			expected: "glesys: some credentials information are missing: GLESYS_API_USER",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				"GLESYS_API_USER": "A",
				"GLESYS_API_KEY":  "",
			},
			expected: "glesys: some credentials information are missing: GLESYS_API_KEY",
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
				require.NotNil(t, p.activeRecords)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiUser  string
		apiKey   string
		expected string
	}{
		{
			desc:    "success",
			apiUser: "A",
			apiKey:  "B",
		},
		{
			desc:     "missing credentials",
			expected: "glesys: incomplete credentials provided",
		},
		{
			desc:     "missing api user",
			apiUser:  "",
			apiKey:   "B",
			expected: "glesys: incomplete credentials provided",
		},
		{
			desc:     "missing api key",
			apiUser:  "A",
			apiKey:   "",
			expected: "glesys: incomplete credentials provided",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("GLESYS_API_USER")
			os.Unsetenv("GLESYS_API_KEY")

			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.APIUser = test.apiUser

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.activeRecords)
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

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
