package mythicbeasts

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUserName,
	EnvPassword).
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
				EnvUserName: "123",
				EnvPassword: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvUserName: "",
				EnvPassword: "",
			},
			expected: "mythicbeasts: some credentials information are missing: MYTHICBEASTS_USERNAME,MYTHICBEASTS_PASSWORD",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvUserName: "",
				EnvPassword: "api_password",
			},
			expected: "mythicbeasts: some credentials information are missing: MYTHICBEASTS_USERNAME",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvUserName: "api_username",
				EnvPassword: "",
			},
			expected: "mythicbeasts: some credentials information are missing: MYTHICBEASTS_PASSWORD",
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
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			username: "api_username",
			password: "api_password",
		},
		{
			desc:     "missing credentials",
			expected: "mythicbeasts: incomplete credentials, missing username and/or password",
		},
		{
			desc:     "missing username",
			username: "",
			password: "api_password",
			expected: "mythicbeasts: incomplete credentials, missing username and/or password",
		},
		{
			desc:     "missing password",
			username: "api_username",
			password: "",
			expected: "mythicbeasts: incomplete credentials, missing username and/or password",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config, err := NewDefaultConfig()
			require.NoError(t, err)

			config.UserName = test.username
			config.Password = test.password

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
