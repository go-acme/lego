package aliesa

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAccessKey,
	EnvSecretKey,
	EnvRAMRole).
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
				EnvAccessKey: "123",
				EnvSecretKey: "456",
			},
		},
		{
			desc: "success (RAM role)",
			envVars: map[string]string{
				EnvRAMRole: "LegoInstanceRole",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAccessKey: "",
				EnvSecretKey: "",
			},
			expected: "aliesa: some credentials information are missing: ALIESA_ACCESS_KEY,ALIESA_SECRET_KEY",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				EnvAccessKey: "",
				EnvSecretKey: "456",
			},
			expected: "aliesa: some credentials information are missing: ALIESA_ACCESS_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAccessKey: "123",
				EnvSecretKey: "",
			},
			expected: "aliesa: some credentials information are missing: ALIESA_SECRET_KEY",
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
		ramRole   string
		apiKey    string
		secretKey string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "123",
			secretKey: "456",
		},
		{
			desc:    "success",
			ramRole: "LegoInstanceRole",
		},
		{
			desc:     "missing credentials",
			expected: "aliesa: ram role or credentials missing",
		},
		{
			desc:      "missing api key",
			secretKey: "456",
			expected:  "aliesa: ram role or credentials missing",
		},
		{
			desc:     "missing secret key",
			apiKey:   "123",
			expected: "aliesa: ram role or credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.SecretKey = test.secretKey
			config.RAMRole = test.ramRole

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
