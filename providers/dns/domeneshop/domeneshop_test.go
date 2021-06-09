package domeneshop

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIToken,
	EnvAPISecret).
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
				EnvAPIToken:  "A",
				EnvAPISecret: "B",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIToken:  "",
				EnvAPISecret: "",
			},
			expected: "domeneshop: some credentials information are missing: DOMENESHOP_API_TOKEN,DOMENESHOP_API_SECRET",
		},
		{
			desc: "missing api token",
			envVars: map[string]string{
				EnvAPIToken:  "",
				EnvAPISecret: "A",
			},
			expected: "domeneshop: some credentials information are missing: DOMENESHOP_API_TOKEN",
		},
		{
			desc: "missing api secret",
			envVars: map[string]string{
				EnvAPIToken:  "A",
				EnvAPISecret: "",
			},
			expected: "domeneshop: some credentials information are missing: DOMENESHOP_API_SECRET",
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
		desc      string
		apiSecret string
		apiToken  string
		expected  string
	}{
		{
			desc:      "success",
			apiToken:  "A",
			apiSecret: "B",
		},
		{
			desc:     "missing credentials",
			expected: "domeneshop: credentials missing",
		},
		{
			desc:      "missing api token",
			apiToken:  "",
			apiSecret: "B",
			expected:  "domeneshop: credentials missing",
		},
		{
			desc:      "missing api secret",
			apiToken:  "A",
			apiSecret: "",
			expected:  "domeneshop: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()

			config.APIToken = test.apiToken
			config.APISecret = test.apiSecret

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
