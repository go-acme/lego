package nicmanager

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvLogin, EnvEmail, EnvPassword, EnvOTP).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success (email)",
			envVars: map[string]string{
				EnvEmail:    "foo@example.com",
				EnvPassword: "secret",
			},
		},
		{
			desc: "success (login.username)",
			envVars: map[string]string{
				EnvLogin:    "foo",
				EnvUsername: "bar",
				EnvPassword: "secret",
			},
		},
		{
			desc:     "missing credentials",
			expected: "nicmanager: some credentials information are missing: NICMANAGER_API_PASSWORD",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvEmail: "foo@example.com",
			},
			expected: "nicmanager: some credentials information are missing: NICMANAGER_API_PASSWORD",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvLogin:    "foo",
				EnvPassword: "secret",
			},
			expected: "nicmanager: credentials missing",
		},
		{
			desc: "missing login",
			envVars: map[string]string{
				EnvUsername: "bar",
				EnvPassword: "secret",
			},
			expected: "nicmanager: credentials missing",
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
		login     string
		username  string
		email     string
		password  string
		otpSecret string
		expected  string
	}{
		{
			desc:     "success (email)",
			email:    "foo@example.com",
			password: "secret",
		},
		{
			desc:     "success (login.username)",
			login:    "john",
			username: "doe",
			password: "secret",
		},
		{
			desc:     "missing credentials",
			expected: "nicmanager: credentials missing",
		},
		{
			desc:     "missing password",
			email:    "foo@example.com",
			expected: "nicmanager: credentials missing",
		},
		{
			desc:     "missing login",
			login:    "",
			username: "doe",
			password: "secret",
			expected: "nicmanager: credentials missing",
		},
		{
			desc:     "missing username",
			login:    "john",
			username: "",
			password: "secret",
			expected: "nicmanager: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Login = test.login
			config.Username = test.username
			config.Email = test.email
			config.Password = test.password
			config.OTPSecret = test.otpSecret

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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
