package glesys

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIUser,
	EnvAPIKey).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUser: "A",
				EnvAPIKey:  "B",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIUser: "",
				EnvAPIKey:  "",
			},
			expected: "glesys: some credentials information are missing: GLESYS_API_USER,GLESYS_API_KEY",
		},
		{
			desc: "missing api user",
			envVars: map[string]string{
				EnvAPIUser: "",
				EnvAPIKey:  "B",
			},
			expected: "glesys: some credentials information are missing: GLESYS_API_USER",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIUser: "A",
				EnvAPIKey:  "",
			},
			expected: "glesys: some credentials information are missing: GLESYS_API_KEY",
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
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.APIUser = test.apiUser

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
