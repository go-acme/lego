package cloudns

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAuthID,
	EnvSubAuthID,
	EnvAuthPassword).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success auth-id",
			envVars: map[string]string{
				EnvAuthID:       "123",
				EnvSubAuthID:    "",
				EnvAuthPassword: "456",
			},
		},
		{
			desc: "success sub-auth-id",
			envVars: map[string]string{
				EnvAuthID:       "",
				EnvSubAuthID:    "123",
				EnvAuthPassword: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAuthID:       "",
				EnvSubAuthID:    "",
				EnvAuthPassword: "",
			},
			expected: "ClouDNS: some credentials information are missing: CLOUDNS_AUTH_ID or CLOUDNS_SUB_AUTH_ID",
		},
		{
			desc: "missing auth-id",
			envVars: map[string]string{
				EnvAuthID:       "",
				EnvSubAuthID:    "",
				EnvAuthPassword: "456",
			},
			expected: "ClouDNS: some credentials information are missing: CLOUDNS_AUTH_ID or CLOUDNS_SUB_AUTH_ID",
		},
		{
			desc: "missing sub-auth-id",
			envVars: map[string]string{
				EnvAuthID:       "",
				EnvSubAuthID:    "",
				EnvAuthPassword: "456",
			},
			expected: "ClouDNS: some credentials information are missing: CLOUDNS_AUTH_ID or CLOUDNS_SUB_AUTH_ID",
		},
		{
			desc: "missing auth-password",
			envVars: map[string]string{
				EnvAuthID:       "123",
				EnvSubAuthID:    "",
				EnvAuthPassword: "",
			},
			expected: "ClouDNS: some credentials information are missing: CLOUDNS_AUTH_PASSWORD",
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
		desc         string
		authID       string
		subAuthID    string
		authPassword string
		expected     string
	}{
		{
			desc:         "success auth-id",
			authID:       "123",
			subAuthID:    "",
			authPassword: "456",
		},
		{
			desc:         "success sub-auth-id",
			authID:       "",
			subAuthID:    "123",
			authPassword: "456",
		},
		{
			desc:     "missing credentials",
			expected: "ClouDNS: credentials missing: authID or subAuthID",
		},
		{
			desc:         "missing auth-id",
			authID:       "",
			subAuthID:    "",
			authPassword: "456",
			expected:     "ClouDNS: credentials missing: authID or subAuthID",
		},
		{
			desc:         "missing sub-auth-id",
			authID:       "",
			subAuthID:    "",
			authPassword: "456",
			expected:     "ClouDNS: credentials missing: authID or subAuthID",
		},
		{
			desc:     "missing auth-password",
			authID:   "123",
			expected: "ClouDNS: credentials missing: authPassword",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthID = test.authID
			config.SubAuthID = test.subAuthID
			config.AuthPassword = test.authPassword

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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
