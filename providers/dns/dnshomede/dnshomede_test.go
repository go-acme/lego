package dnshomede

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvCredentials).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvCredentials: "example.org:123",
			},
		},
		{
			desc: "success multiple domains",
			envVars: map[string]string{
				EnvCredentials: "example.org:123,example.com:456,example.net:789",
			},
		},
		{
			desc: "invalid credentials",
			envVars: map[string]string{
				EnvCredentials: ",",
			},
			expected: `dnshomede: credentials: incorrect pair: `,
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvCredentials: "example.org:",
			},
			expected: `dnshomede: missing password: "example.org:"`,
		},
		{
			desc: "missing domain",
			envVars: map[string]string{
				EnvCredentials: ":123",
			},
			expected: `dnshomede: missing domain: ":123"`,
		},
		{
			desc: "invalid credentials, partial",
			envVars: map[string]string{
				EnvCredentials: "example.org:123,example.net",
			},
			expected: "dnshomede: credentials: incorrect pair: example.net",
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvCredentials: "",
			},
			expected: "dnshomede: some credentials information are missing: DNSHOMEDE_CREDENTIALS",
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
			expected: "dnshomede: missing credentials",
		},
		{
			desc:     "missing domain",
			creds:    map[string]string{"": "123"},
			expected: `dnshomede: missing domain: ":123"`,
		},
		{
			desc:     "missing password",
			creds:    map[string]string{"example.org": ""},
			expected: `dnshomede: missing password: "example.org:"`,
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
