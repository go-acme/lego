package sonic

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvUserID).
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
				EnvUserID: "dummy",
				EnvAPIKey: "dummy",
			},
		},
		{
			desc:     "missing all credentials",
			envVars:  map[string]string{},
			expected: "sonic: some credentials information are missing: SONIC_USER_ID,SONIC_API_KEY",
		},
		{
			desc: "no userid",
			envVars: map[string]string{
				EnvAPIKey: "dummy",
			},
			expected: "sonic: some credentials information are missing: SONIC_USER_ID",
		},
		{
			desc: "no apikey",
			envVars: map[string]string{
				EnvUserID: "dummy",
			},
			expected: `sonic: some credentials information are missing: SONIC_API_KEY`,
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		userID   string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			userID: "dummy",
			apiKey: "dummy",
		},
		{
			desc:     "missing all credentials",
			expected: "sonic: credentials are missing",
		},
		{
			desc:     "missing userid",
			apiKey:   "dummy",
			expected: "sonic: credentials are missing",
		},
		{
			desc:     "missing apikey",
			userID:   "dummy",
			expected: "sonic: credentials are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.UserID = test.userID
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
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
