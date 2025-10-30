package manageengine

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvClientID, EnvClientSecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvClientID:     "abc",
				EnvClientSecret: "secret",
			},
		},
		{
			desc: "missing client ID",
			envVars: map[string]string{
				EnvClientID:     "",
				EnvClientSecret: "secret",
			},
			expected: "manageengine: some credentials information are missing: MANAGEENGINE_CLIENT_ID",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				EnvClientID:     "abc",
				EnvClientSecret: "",
			},
			expected: "manageengine: some credentials information are missing: MANAGEENGINE_CLIENT_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "manageengine: some credentials information are missing: MANAGEENGINE_CLIENT_ID,MANAGEENGINE_CLIENT_SECRET",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc         string
		clientID     string
		clientSecret string
		expected     string
	}{
		{
			desc:         "success",
			clientID:     "abc",
			clientSecret: "secret",
		},
		{
			desc:         "missing client ID",
			clientSecret: "secret",
			expected:     "manageengine: credentials missing",
		},
		{
			desc:     "missing client secret",
			clientID: "abc",
			expected: "manageengine: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "manageengine: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ClientID = test.clientID
			config.ClientSecret = test.clientSecret

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
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
