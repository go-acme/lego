package luadns

import (
	"testing"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/go-acme/lego/v3/providers/dns/luadns/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIUsername,
	EnvAPIToken).
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
				EnvAPIUsername: "123",
				EnvAPIToken:    "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIUsername: "",
				EnvAPIToken:    "",
			},
			expected: "luadns: some credentials information are missing: LUADNS_API_USERNAME,LUADNS_API_TOKEN",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAPIUsername: "",
				EnvAPIToken:    "456",
			},
			expected: "luadns: some credentials information are missing: LUADNS_API_USERNAME",
		},
		{
			desc: "missing api token",
			envVars: map[string]string{
				EnvAPIUsername: "123",
				EnvAPIToken:    "",
			},
			expected: "luadns: some credentials information are missing: LUADNS_API_TOKEN",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider(nil)

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
		tll       int
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "123",
			apiSecret: "456",
			tll:       minTTL,
		},
		{
			desc:     "missing credentials",
			tll:      minTTL,
			expected: "luadns: credentials missing",
		},
		{
			desc:      "missing username",
			apiSecret: "456",
			tll:       minTTL,
			expected:  "luadns: credentials missing",
		},
		{
			desc:     "missing api token",
			apiKey:   "123",
			tll:      minTTL,
			expected: "luadns: credentials missing",
		},
		{
			desc:      "invalid TTL",
			apiKey:    "123",
			apiSecret: "456",
			tll:       30,
			expected:  "luadns: invalid TTL, TTL (30) must be greater than 300",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.APIUsername = test.apiKey
			config.APIToken = test.apiSecret
			config.TTL = test.tll

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

func TestDNSProvider_findZone(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		zones    []internal.DNSZone
		expected *internal.DNSZone
	}{
		{
			desc:   "simple domain",
			domain: "example.org",
			zones: []internal.DNSZone{
				{Name: "example.org"},
				{Name: "example.com"},
			},
			expected: &internal.DNSZone{Name: "example.org"},
		},
		{
			desc:   "sub domain",
			domain: "aaa.example.org",
			zones: []internal.DNSZone{
				{Name: "example.org"},
				{Name: "aaa.example.org"},
				{Name: "bbb.example.org"},
				{Name: "example.com"},
			},
			expected: &internal.DNSZone{Name: "aaa.example.org"},
		},
		{
			desc:   "empty zone name",
			domain: "example.org",
			zones: []internal.DNSZone{
				{},
			},
		},
		{
			desc:   "not found",
			domain: "example.org",
			zones: []internal.DNSZone{
				{Name: "example.net"},
				{Name: "aaa.example.net"},
				{Name: "bbb.example.net"},
				{Name: "example.com"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone := findZone(test.zones, test.domain)
			assert.Equal(t, test.expected, zone)
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
