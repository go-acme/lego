package googledomains

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAccessToken).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvAccessToken, envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAccessToken: "abc",
			},
			expected: "",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "googledomains: some credentials information are missing: GOOGLE_DOMAINS_ACCESS_TOKEN",
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
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		accessToken string
		expected    string
	}{
		{
			desc:        "success",
			accessToken: "abc",
		},
		{
			desc:     "missing credentials",
			expected: "googledomains: access token is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessToken = test.accessToken

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
