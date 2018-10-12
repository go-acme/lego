package ns1

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	liveTest      bool
	envTestAPIKey string
	envTestDomain string
)

func init() {
	envTestAPIKey = os.Getenv("NS1_API_KEY")
	envTestDomain = os.Getenv("NS1_DOMAIN")
	if len(envTestAPIKey) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("NS1_API_KEY", envTestAPIKey)
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
				"NS1_API_KEY": "123",
			},
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				"NS1_API_KEY": "",
			},
			expected: "ns1: some credentials information are missing: NS1_API_KEY",
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
		desc     string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
		},
		{
			desc:     "missing credentials",
			expected: "ns1: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("NS1_API_KEY")

			config := NewDefaultConfig()
			config.APIKey = test.apiKey

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

func Test_getAuthZone(t *testing.T) {
	type expected struct {
		AuthZone string
		Error    string
	}

	testCases := []struct {
		desc     string
		fqdn     string
		expected expected
	}{
		{
			desc: "valid fqdn",
			fqdn: "_acme-challenge.myhost.sub.example.com.",
			expected: expected{
				AuthZone: "example.com",
			},
		},
		{
			desc: "invalid fqdn",
			fqdn: "_acme-challenge.myhost.sub.example.com",
			expected: expected{
				Error: "dns: domain must be fully qualified",
			},
		},
		{
			desc: "invalid authority",
			fqdn: "_acme-challenge.myhost.sub.domain.tld.",
			expected: expected{
				Error: "could not find the start of authority",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			authZone, err := getAuthZone(test.fqdn)

			if len(test.expected.Error) > 0 {
				assert.EqualError(t, err, test.expected.Error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.AuthZone, authZone)
			}
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
