package mailinabox

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvBaseURL, EnvEmail, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvBaseURL:  "https://example.com",
				EnvEmail:    "user@example.com",
				EnvPassword: "secret",
			},
		},
		{
			desc: "missing base URL",
			envVars: map[string]string{
				EnvEmail:    "user@example.com",
				EnvPassword: "secret",
			},
			expected: "mailinabox: some credentials information are missing: MAILINABOX_BASE_URL",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				EnvBaseURL:  "https://example.com",
				EnvPassword: "secret",
			},
			expected: "mailinabox: some credentials information are missing: MAILINABOX_EMAIL",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvBaseURL: "https://example.com",
				EnvEmail:   "user@example.com",
			},
			expected: "mailinabox: some credentials information are missing: MAILINABOX_PASSWORD",
		},
		{
			desc:     "missing all options",
			expected: "mailinabox: some credentials information are missing: MAILINABOX_BASE_URL,MAILINABOX_EMAIL,MAILINABOX_PASSWORD",
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
		email    string
		password string
		expected string
	}{
		{
			desc:     "success",
			baseURL:  "https://example.com",
			email:    "user@example.com",
			password: "secret",
		},
		{
			desc:     "missing base URL",
			email:    "user@example.com",
			password: "secret",
			expected: "mailinabox: missing base URL",
		},
		{
			desc:     "missing email",
			baseURL:  "https://example.com",
			password: "secret",
			expected: "mailinabox: incomplete credentials, missing email or password",
		},
		{
			desc:     "missing password",
			baseURL:  "https://example.com",
			email:    "user@example.com",
			expected: "mailinabox: incomplete credentials, missing email or password",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = test.baseURL
			config.Email = test.email
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
