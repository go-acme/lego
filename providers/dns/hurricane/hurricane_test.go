package hurricane

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvTokens).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvTokens: "example.org:123",
			},
		},
		{
			desc: "success multiple domains",
			envVars: map[string]string{
				EnvTokens: "example.org:123,example.com:456,example.net:789",
			},
		},
		{
			desc: "invalid credentials",
			envVars: map[string]string{
				EnvTokens: ",",
			},
			expected: "hurricane: incorrect credential pair: ",
		},
		{
			desc: "invalid credentials, partial",
			envVars: map[string]string{
				EnvTokens: "example.org:123,example.net",
			},
			expected: "hurricane: incorrect credential pair: example.net",
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvTokens: "",
			},
			expected: "hurricane: some credentials information are missing: HURRICANE_TOKENS",
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
		creds    map[string]string
		expected string
	}{
		{
			desc:  "success",
			creds: map[string]string{"example.org": "123"},
		},
		{
			desc: "success multiple domains",
			creds: map[string]string{
				"example.org": "123",
				"example.com": "456",
				"example.net": "789",
			},
		},
		{
			desc:     "missing credentials",
			expected: "hurricane: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Credentials = test.creds

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
