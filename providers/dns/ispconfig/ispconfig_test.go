package ispconfig

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvServerURL,
	EnvUsername,
	EnvPassword,
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
				EnvServerURL: "https://example.com:80/",
				EnvUsername:  "user",
				EnvPassword:  "secret",
			},
		},
		{
			desc: "missing server URL",
			envVars: map[string]string{
				EnvServerURL: "",
				EnvUsername:  "user",
				EnvPassword:  "secret",
			},
			expected: "ispconfig: some credentials information are missing: ISPCONFIG_SERVER_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvServerURL: "https://example.com:80/",
				EnvUsername:  "",
				EnvPassword:  "secret",
			},
			expected: "ispconfig: some credentials information are missing: ISPCONFIG_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvServerURL: "https://example.com:80/",
				EnvUsername:  "user",
				EnvPassword:  "",
			},
			expected: "ispconfig: some credentials information are missing: ISPCONFIG_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "ispconfig: some credentials information are missing: ISPCONFIG_SERVER_URL,ISPCONFIG_USERNAME,ISPCONFIG_PASSWORD",
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
		desc      string
		serverURL string
		username  string
		password  string
		expected  string
	}{
		{
			desc:      "success",
			serverURL: "https://example.com:80/",
			username:  "user",
			password:  "secret",
		},
		{
			desc:     "missing server URL",
			username: "user",
			password: "secret",
			expected: "ispconfig: missing server URL",
		},
		{
			desc:      "missing username",
			serverURL: "https://example.com:80/",
			password:  "secret",
			expected:  "ispconfig: credentials missing",
		},
		{
			desc:      "missing password",
			serverURL: "https://example.com:80/",
			username:  "user",
			expected:  "ispconfig: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "ispconfig: missing server URL",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ServerURL = test.serverURL
			config.Username = test.username
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
