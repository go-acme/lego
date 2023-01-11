package ultradns

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUsername,
	EnvPassword,
	EnvEndpoint,
	EnvTTL,
	EnvPropagationTimeout,
	EnvPollingInterval).
	WithDomain(envDomain)

func TestNewDefaultConfig(t *testing.T) {
	defer envTest.RestoreEnv()

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected *Config
	}{
		{
			desc: "default configuration",
			expected: &Config{
				Endpoint:           "https://api.ultradns.com/",
				TTL:                120,
				PropagationTimeout: 2 * time.Minute,
				PollingInterval:    4 * time.Second,
			},
		},
		{
			desc: "input configuration",
			envVars: map[string]string{
				EnvEndpoint:           "https://example.com/",
				EnvTTL:                "99",
				EnvPropagationTimeout: "60",
				EnvPollingInterval:    "60",
			},
			expected: &Config{
				Endpoint:           "https://example.com/",
				TTL:                99,
				PropagationTimeout: 60 * time.Second,
				PollingInterval:    60 * time.Second,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			envTest.ClearEnv()
			envTest.Apply(test.envVars)

			config := NewDefaultConfig()

			assert.Equal(t, test.expected, config)
		})
	}
}

func TestNewDNSProvider(t *testing.T) {
	defer envTest.RestoreEnv()

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc:     "missing username and password",
			expected: "ultradns: some credentials information are missing: ULTRADNS_USERNAME,ULTRADNS_PASSWORD",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvPassword: "password",
			},
			expected: "ultradns: some credentials information are missing: ULTRADNS_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "username",
			},
			expected: "ultradns: some credentials information are missing: ULTRADNS_PASSWORD",
		},
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "username",
				EnvPassword: "password",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			envTest.ClearEnv()
			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
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
			expected: "ultradns: config validation failure: username is missing",
		},
		{
			desc:     "missing username",
			username: "",
			password: "api_password",
			expected: "ultradns: config validation failure: username is missing",
		},
		{
			desc:     "missing password",
			username: "api_username",
			password: "",
			expected: "ultradns: config validation failure: password is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
