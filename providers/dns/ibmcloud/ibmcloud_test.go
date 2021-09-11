package ibmcloud

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvAPIKey).
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
				EnvUsername: "123",
				EnvAPIKey:   "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvUsername: "",
				EnvAPIKey:   "",
			},
			expected: "ibmcloud: some credentials information are missing: SOFTLAYER_USERNAME,SOFTLAYER_API_KEY",
		},
		{
			desc: "missing access token",
			envVars: map[string]string{
				EnvUsername: "",
				EnvAPIKey:   "456",
			},
			expected: "ibmcloud: some credentials information are missing: SOFTLAYER_USERNAME",
		},
		{
			desc: "missing token secret",
			envVars: map[string]string{
				EnvUsername: "123",
				EnvAPIKey:   "",
			},
			expected: "ibmcloud: some credentials information are missing: SOFTLAYER_API_KEY",
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
				require.NotNil(t, p.wrapper)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		username string
		apiKey   string
		expected string
	}{
		{
			desc:     "success",
			username: "123",
			apiKey:   "456",
		},
		{
			desc:     "missing credentials",
			expected: "ibmcloud: username is missing",
		},
		{
			desc:     "missing token",
			apiKey:   "456",
			expected: "ibmcloud: username is missing",
		},
		{
			desc:     "missing secret",
			username: "123",
			expected: "ibmcloud: API key is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.wrapper)
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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
