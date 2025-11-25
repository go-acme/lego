package edgeone

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvSecretID,
	EnvSecretKey,
	EnvZonesMapping,
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
				EnvSecretID:  "123",
				EnvSecretKey: "456",
			},
		},
		{
			desc: "success with zones mapping",
			envVars: map[string]string{
				EnvSecretID:     "123",
				EnvSecretKey:    "456",
				EnvZonesMapping: "example.org:id1,example.com:id2",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvSecretID:  "",
				EnvSecretKey: "",
			},
			expected: "edgeone: some credentials information are missing: EDGEONE_SECRET_ID,EDGEONE_SECRET_KEY",
		},
		{
			desc: "missing access id",
			envVars: map[string]string{
				EnvSecretID:  "",
				EnvSecretKey: "456",
			},
			expected: "edgeone: some credentials information are missing: EDGEONE_SECRET_ID",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvSecretID:  "123",
				EnvSecretKey: "",
			},
			expected: "edgeone: some credentials information are missing: EDGEONE_SECRET_KEY",
		},
		{
			desc: "invalid mapping",
			envVars: map[string]string{
				EnvSecretID:     "123",
				EnvSecretKey:    "456",
				EnvZonesMapping: "example.org:id1,example.com",
			},
			expected: "edgeone: zones mapping: incorrect pair: example.com",
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
		secretID  string
		secretKey string
		expected  string
	}{
		{
			desc:      "success",
			secretID:  "123",
			secretKey: "456",
		},
		{
			desc:     "missing credentials",
			expected: "edgeone: credentials missing",
		},
		{
			desc:      "missing secret id",
			secretKey: "456",
			expected:  "edgeone: credentials missing",
		},
		{
			desc:     "missing secret key",
			secretID: "123",
			expected: "edgeone: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.SecretID = test.secretID
			config.SecretKey = test.secretKey

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
