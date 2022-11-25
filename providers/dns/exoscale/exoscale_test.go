package exoscale

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPISecret,
	EnvAPIKey).
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
				EnvAPIKey:    "123",
				EnvAPISecret: "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvAPISecret: "",
			},
			expected: "exoscale: some credentials information are missing: EXOSCALE_API_KEY,EXOSCALE_API_SECRET",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvAPISecret: "456",
			},
			expected: "exoscale: some credentials information are missing: EXOSCALE_API_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAPIKey:    "123",
				EnvAPISecret: "",
			},
			expected: "exoscale: some credentials information are missing: EXOSCALE_API_SECRET",
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
		apiKey    string
		apiSecret string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "123",
			apiSecret: "456",
		},
		{
			desc:     "missing credentials",
			expected: "exoscale: credentials missing",
		},
		{
			desc:      "missing api key",
			apiSecret: "456",
			expected:  "exoscale: credentials missing",
		},
		{
			desc:     "missing secret key",
			apiKey:   "123",
			expected: "exoscale: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.APISecret = test.apiSecret

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

func TestDNSProvider_FindZoneAndRecordName(t *testing.T) {
	config := NewDefaultConfig()
	config.APIKey = "example@example.com"
	config.APISecret = "123"

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	type expected struct {
		zone       string
		recordName string
	}

	testCases := []struct {
		desc     string
		fqdn     string
		expected expected
	}{
		{
			desc: "Extract root record name",
			fqdn: "_acme-challenge.bar.com.",
			expected: expected{
				zone:       "bar.com",
				recordName: "_acme-challenge",
			},
		},
		{
			desc: "Extract sub record name",
			fqdn: "_acme-challenge.foo.bar.com.",
			expected: expected{
				zone:       "bar.com",
				recordName: "_acme-challenge.foo",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, recordName, err := provider.findZoneAndRecordName(test.fqdn)
			require.NoError(t, err)
			assert.Equal(t, test.expected.zone, zone)
			assert.Equal(t, test.expected.recordName, recordName)
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

	// Present Twice to handle create / update
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
