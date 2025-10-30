package clouddns

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvClientID,
	EnvEmail,
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
				EnvClientID: "client123",
				EnvEmail:    "test@example.com",
				EnvPassword: "password123",
			},
		},
		{
			desc: "missing clientId",
			envVars: map[string]string{
				EnvClientID: "",
				EnvEmail:    "test@example.com",
				EnvPassword: "password123",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_CLIENT_ID",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				EnvClientID: "client123",
				EnvEmail:    "",
				EnvPassword: "password123",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_EMAIL",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvClientID: "client123",
				EnvEmail:    "test@example.com",
				EnvPassword: "",
			},
			expected: "clouddns: some credentials information are missing: CLOUDDNS_PASSWORD",
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
		clientID string
		email    string
		password string
		expected string
	}{
		{
			desc:     "success",
			clientID: "ID",
			email:    "test@example.com",
			password: "secret",
		},
		{
			desc:     "missing credentials",
			expected: "clouddns: credentials missing",
		},
		{
			desc:     "missing client ID",
			clientID: "",
			email:    "test@example.com",
			password: "secret",
			expected: "clouddns: credentials missing",
		},
		{
			desc:     "missing email",
			clientID: "ID",
			email:    "",
			password: "secret",
			expected: "clouddns: credentials missing",
		},
		{
			desc:     "missing password",
			clientID: "ID",
			email:    "test@example.com",
			password: "",
			expected: "clouddns: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ClientID = test.clientID
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
