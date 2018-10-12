package exoscale

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	liveTest         bool
	envTestAPIKey    string
	envTestAPISecret string
	envTestDomain    string
)

func init() {
	envTestAPISecret = os.Getenv("EXOSCALE_API_SECRET")
	envTestAPIKey = os.Getenv("EXOSCALE_API_KEY")
	envTestDomain = os.Getenv("EXOSCALE_DOMAIN")

	if len(envTestAPIKey) > 0 && len(envTestAPISecret) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("EXOSCALE_API_KEY", envTestAPIKey)
	os.Setenv("EXOSCALE_API_SECRET", envTestAPISecret)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"EXOSCALE_API_KEY":    "123",
				"EXOSCALE_API_SECRET": "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"EXOSCALE_API_KEY":    "",
				"EXOSCALE_API_SECRET": "",
			},
			expected: "exoscale: some credentials information are missing: EXOSCALE_API_KEY,EXOSCALE_API_SECRET",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				"EXOSCALE_API_KEY":    "",
				"EXOSCALE_API_SECRET": "456",
			},
			expected: "exoscale: some credentials information are missing: EXOSCALE_API_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				"EXOSCALE_API_KEY":    "123",
				"EXOSCALE_API_SECRET": "",
			},
			expected: "exoscale: some credentials information are missing: EXOSCALE_API_SECRET",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			for key, value := range test.envVars {
				if len(value) == 0 {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
			defer restoreEnv()
			os.Unsetenv("EXOSCALE_API_KEY")
			os.Unsetenv("EXOSCALE_API_SECRET")

			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.APISecret = test.apiSecret

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
		domain   string
		expected expected
	}{
		{
			desc:   "Extract root record name",
			fqdn:   "_acme-challenge.bar.com.",
			domain: "bar.com",
			expected: expected{
				zone:       "bar.com",
				recordName: "_acme-challenge",
			},
		},
		{
			desc:   "Extract sub record name",
			fqdn:   "_acme-challenge.foo.bar.com.",
			domain: "foo.bar.com",
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

			zone, recordName, err := provider.FindZoneAndRecordName(test.fqdn, test.domain)
			require.NoError(t, err)
			assert.Equal(t, test.expected.zone, zone)
			assert.Equal(t, test.expected.recordName, recordName)
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)

	// Present Twice to handle create / update
	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
