package luadns

import (
	"testing"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIUsername,
	EnvAPIToken).
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
				EnvAPIUsername: "123",
				EnvAPIToken:    "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIUsername: "",
				EnvAPIToken:    "",
			},
			expected: "luadns: some credentials information are missing: LUADNS_API_USERNAME,LUADNS_API_TOKEN",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAPIUsername: "",
				EnvAPIToken:    "456",
			},
			expected: "luadns: some credentials information are missing: LUADNS_API_USERNAME",
		},
		{
			desc: "missing api token",
			envVars: map[string]string{
				EnvAPIUsername: "123",
				EnvAPIToken:    "",
			},
			expected: "luadns: some credentials information are missing: LUADNS_API_TOKEN",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
		desc      string
		apiKey    string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "123",
			apiSecret: "456",
		},
		{
			desc:     "missing credentials",
			expected: "luadns: credentials missing",
		},
		{
			desc:      "missing username",
			apiSecret: "456",
			expected:  "luadns: credentials missing",
		},
		{
			desc:     "missing api token",
			apiKey:   "123",
			expected: "luadns: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUsername = test.apiKey
			config.APIToken = test.apiSecret

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
