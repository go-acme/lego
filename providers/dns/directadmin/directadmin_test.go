package directadmin

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIURL, EnvUsername, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIURL:   "https://example.com:2222",
				EnvUsername: "test",
				EnvPassword: "secret",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "directadmin: some credentials information are missing: DIRECTADMIN_API_URL,DIRECTADMIN_USERNAME,DIRECTADMIN_PASSWORD",
		},
		{
			desc: "missing API URL",
			envVars: map[string]string{
				EnvUsername: "test",
				EnvPassword: "secret",
			},
			expected: "directadmin: some credentials information are missing: DIRECTADMIN_API_URL",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAPIURL:   "https://example.com:2222",
				EnvPassword: "secret",
			},
			expected: "directadmin: some credentials information are missing: DIRECTADMIN_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvAPIURL:   "https://example.com:2222",
				EnvUsername: "test",
			},
			expected: "directadmin: some credentials information are missing: DIRECTADMIN_PASSWORD",
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
				require.NotNil(t, p.client)
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
			username: "test",
			password: "secret",
		},
		{
			desc:     "missing API URL",
			expected: "directadmin: missing API URL",
		},
		{
			desc:     "missing username",
			baseURL:  "https://example.com",
			expected: "directadmin: some credentials information are missing",
		},
		{
			desc:     "missing password",
			baseURL:  "https://example.com",
			username: "test",
			expected: "directadmin: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = test.baseURL
			config.Username = test.username
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.client)
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
