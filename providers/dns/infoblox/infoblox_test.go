package infoblox

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvHost,
	EnvPort,
	EnvUsername,
	EnvPassword,
	EnvSSLVerify,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvHost:      "example.com",
				EnvUsername:  "user",
				EnvPassword:  "secret",
				EnvSSLVerify: "false",
			},
		},
		{
			desc: "missing host",
			envVars: map[string]string{
				EnvHost:      "",
				EnvUsername:  "user",
				EnvPassword:  "secret",
				EnvSSLVerify: "false",
			},
			expected: "infoblox: some credentials information are missing: INFOBLOX_HOST",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvHost:      "example.com",
				EnvUsername:  "",
				EnvPassword:  "secret",
				EnvSSLVerify: "false",
			},
			expected: "infoblox: some credentials information are missing: INFOBLOX_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvHost:      "example.com",
				EnvUsername:  "user",
				EnvPassword:  "",
				EnvSSLVerify: "false",
			},
			expected: "infoblox: some credentials information are missing: INFOBLOX_PASSWORD",
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
		host     string
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			host:     "example.com",
			username: "user",
			password: "secret",
		},
		{
			desc:     "missing host",
			host:     "",
			username: "user",
			password: "secret",
			expected: "infoblox: missing host",
		},
		{
			desc:     "missing username",
			host:     "example.com",
			username: "",
			password: "secret",
			expected: "infoblox: missing credentials",
		},
		{
			desc:     "missing password",
			host:     "example.com",
			username: "user",
			password: "",
			expected: "infoblox: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Host = test.host
			config.Username = test.username
			config.Password = test.password
			config.SSLVerify = false

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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
