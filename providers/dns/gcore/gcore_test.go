package gcore

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvPermanentAPIToken).WithDomain(envNamespace + "DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvPermanentAPIToken: "A",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvPermanentAPIToken: "",
			},
			expected: "gcore: some credentials information are missing: GCORE_PERMANENT_API_TOKEN",
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
		desc     string
		apiToken string
		expected string
	}{
		{
			desc:     "success",
			apiToken: "A",
		},
		{
			desc:     "missing credentials",
			expected: "gcore: incomplete credentials provided",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIToken = test.apiToken

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

func Test_extractAllZones(t *testing.T) {
	testCases := []struct {
		desc     string
		fqdn     string
		expected []string
	}{
		{
			desc:     "success",
			fqdn:     "_acme-challenge.my.test.domain.com.",
			expected: []string{"my.test.domain.com", "test.domain.com", "domain.com"},
		},
		{
			desc: "empty",
			fqdn: "_acme-challenge.com.",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := extractAllZones(test.fqdn)
			assert.Equal(t, test.expected, got)
		})
	}
}
