package allinkl

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvLogin, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvLogin:    "user",
				EnvPassword: "secret",
			},
		},
		{
			desc: "missing credentials: account name",
			envVars: map[string]string{
				EnvLogin:    "",
				EnvPassword: "secret",
			},
			expected: "allinkl: some credentials information are missing: ALL_INKL_LOGIN",
		},
		{
			desc: "missing credentials: api key",
			envVars: map[string]string{
				EnvLogin:    "user",
				EnvPassword: "",
			},
			expected: "allinkl: some credentials information are missing: ALL_INKL_PASSWORD",
		},
		{
			desc: "missing credentials: all",
			envVars: map[string]string{
				EnvLogin:    "",
				EnvPassword: "",
			},
			expected: "allinkl: some credentials information are missing: ALL_INKL_LOGIN,ALL_INKL_PASSWORD",
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
		desc     string
		login    string
		password string
		expected string
	}{
		{
			desc:     "success",
			login:    "user",
			password: "secret",
		},
		{
			desc:     "missing account name",
			password: "secret",
			expected: "allinkl: missing credentials",
		},
		{
			desc:     "missing api key",
			login:    "user",
			expected: "allinkl: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Login = test.login
			config.Password = test.password

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
