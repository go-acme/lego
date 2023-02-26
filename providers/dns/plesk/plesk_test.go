package plesk

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvServerBaseURL,
	EnvUsername,
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
				EnvServerBaseURL: "https//example.com",
				EnvUsername:      "user",
				EnvPassword:      "secret",
			},
		},
		{
			desc: "missing server base URL",
			envVars: map[string]string{
				EnvServerBaseURL: "",
				EnvUsername:      "user",
				EnvPassword:      "secret",
			},
			expected: "plesk: some credentials information are missing: PLESK_SERVER_BASE_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvServerBaseURL: "https//example.com",
				EnvUsername:      "",
				EnvPassword:      "secret",
			},
			expected: "plesk: some credentials information are missing: PLESK_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvServerBaseURL: "https//example.com",
				EnvUsername:      "user",
				EnvPassword:      "",
			},
			expected: "plesk: some credentials information are missing: PLESK_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "plesk: some credentials information are missing: PLESK_SERVER_BASE_URL,PLESK_USERNAME,PLESK_PASSWORD",
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
		baseURL  string
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			baseURL:  "https://example.com",
			username: "user",
			password: "secret",
		},
		{
			desc:     "missing base URL",
			username: "user",
			password: "secret",
			expected: "plesk: missing server base URL",
		},
		{
			desc:     "missing username",
			baseURL:  "https://example.com",
			password: "secret",
			expected: "plesk: incomplete credentials, missing username and/or password",
		},
		{
			desc:     "missing password",
			baseURL:  "https://example.com",
			username: "user",
			expected: "plesk: incomplete credentials, missing username and/or password",
		},
		{
			desc:     "missing credential",
			baseURL:  "https://example.com",
			expected: "plesk: incomplete credentials, missing username and/or password",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.baseURL = test.baseURL
			config.Username = test.username
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
