package scannet

import (
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvClientID, EnvClientSecret).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success (deprecated)",
			envVars: map[string]string{
				EnvAPIKey: "secret",
			},
		},
		{
			desc: "success",
			envVars: map[string]string{
				EnvClientID:     "xxx",
				EnvClientSecret: "secret",
			},
		},
		{
			desc: "missing client ID",
			envVars: map[string]string{
				EnvClientID:     "",
				EnvClientSecret: "secret",
			},
			expected: "scannet: some credentials information are missing: SCANNET_CLIENT_ID",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				EnvClientID:     "xxx",
				EnvClientSecret: "",
			},
			expected: "scannet: some credentials information are missing: SCANNET_CLIENT_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "scannet: some credentials information are missing: SCANNET_CLIENT_ID,SCANNET_CLIENT_SECRET",
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
				require.NotNil(t, p.prv)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc         string
		apiKey       string
		clientID     string
		clientSecret string
		expected     string
	}{
		{
			desc:   "success (deprecated)",
			apiKey: "secret",
		},
		{
			desc:         "success",
			clientID:     "xxx",
			clientSecret: "secret",
		},
		{
			desc:         "missing client ID",
			clientSecret: "secret",
			expected:     "scannet: credentials missing",
		},
		{
			desc:     "missing client secret",
			clientID: "xxx",
			expected: "scannet: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "scannet: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.ClientID = test.clientID
			config.ClientSecret = test.clientSecret

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.prv)
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

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
