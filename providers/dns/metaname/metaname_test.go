package metaname

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAccountReference, EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAccountReference: "user",
				EnvAPIKey:           "secret",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAccountReference: "",
				EnvAPIKey:           "secret",
			},
			expected: "metaname: some credentials information are missing: METANAME_ACCOUNT_REFERENCE",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvAccountReference: "user",
				EnvAPIKey:           "",
			},
			expected: "metaname: some credentials information are missing: METANAME_API_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "metaname: some credentials information are missing: METANAME_ACCOUNT_REFERENCE,METANAME_API_KEY",
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
		desc             string
		accountReference string
		apiKey           string
		expected         string
	}{
		{
			desc:             "success",
			accountReference: "user",
			apiKey:           "secret",
		},
		{
			desc:     "missing username",
			apiKey:   "secret",
			expected: "metaname: missing account reference",
		},
		{
			desc:             "missing password",
			accountReference: "user",
			expected:         "metaname: missing api key",
		},
		{
			desc:     "missing all",
			expected: "metaname: missing account reference",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()

			config.AccountReference = test.accountReference
			config.APIKey = test.apiKey

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
