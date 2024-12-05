package bunny

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/nrdcg/bunny-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
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
				EnvAPIKey: "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "bunny: some credentials information are missing: BUNNY_API_KEY",
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
		desc     string
		apiKey   string
		ttl      int
		expected string
	}{
		{
			desc:   "success",
			ttl:    minTTL,
			apiKey: "123",
		},
		{
			desc:     "missing credentials",
			ttl:      minTTL,
			expected: "bunny: credentials missing",
		},
		{
			desc:     "invalid TTL",
			apiKey:   "123",
			ttl:      10,
			expected: "bunny: invalid TTL, TTL (10) must be greater than 60",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.TTL = test.ttl

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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func Test_findZone(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		items    []*bunny.DNSZone
		expected *bunny.DNSZone
	}{
		{
			desc:   "found subdomain",
			domain: "_acme-challenge.foo.bar.example.com",
			items: []*bunny.DNSZone{
				{ID: pointer[int64](1), Domain: pointer("example.com")},
				{ID: pointer[int64](2), Domain: pointer("example.org")},
				{ID: pointer[int64](4), Domain: pointer("bar.example.org")},
				{ID: pointer[int64](5), Domain: pointer("bar.example.com")},
				{ID: pointer[int64](6), Domain: pointer("foo.example.com")},
			},
			expected: &bunny.DNSZone{
				ID:     pointer[int64](5),
				Domain: pointer("bar.example.com"),
			},
		},
		{
			desc:   "found apex",
			domain: "_acme-challenge.foo.bar.example.com",
			items: []*bunny.DNSZone{
				{ID: pointer[int64](1), Domain: pointer("example.com")},
				{ID: pointer[int64](2), Domain: pointer("example.org")},
				{ID: pointer[int64](4), Domain: pointer("bar.example.org")},
				{ID: pointer[int64](6), Domain: pointer("foo.example.com")},
			},
			expected: &bunny.DNSZone{
				ID:     pointer[int64](1),
				Domain: pointer("example.com"),
			},
		},
		{
			desc:   "not found",
			domain: "_acme-challenge.foo.bar.example.com",
			items: []*bunny.DNSZone{
				{ID: pointer[int64](2), Domain: pointer("example.org")},
				{ID: pointer[int64](4), Domain: pointer("bar.example.org")},
				{ID: pointer[int64](6), Domain: pointer("foo.example.com")},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zones := &bunny.DNSZones{Items: test.items}

			zone := findZone(zones, test.domain)

			assert.Equal(t, test.expected, zone)
		})
	}
}
