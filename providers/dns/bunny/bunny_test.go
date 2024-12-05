package bunny

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
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
				{ID: ptr.Pointer[int64](1), Domain: ptr.Pointer("example.com")},
				{ID: ptr.Pointer[int64](2), Domain: ptr.Pointer("example.org")},
				{ID: ptr.Pointer[int64](4), Domain: ptr.Pointer("bar.example.org")},
				{ID: ptr.Pointer[int64](5), Domain: ptr.Pointer("bar.example.com")},
				{ID: ptr.Pointer[int64](6), Domain: ptr.Pointer("foo.example.com")},
			},
			expected: &bunny.DNSZone{
				ID:     ptr.Pointer[int64](5),
				Domain: ptr.Pointer("bar.example.com"),
			},
		},
		{
			desc:   "found the longest subdomain",
			domain: "_acme-challenge.foo.bar.example.com",
			items: []*bunny.DNSZone{
				{ID: ptr.Pointer[int64](7), Domain: ptr.Pointer("foo.bar.example.com")},
				{ID: ptr.Pointer[int64](1), Domain: ptr.Pointer("example.com")},
				{ID: ptr.Pointer[int64](2), Domain: ptr.Pointer("example.org")},
				{ID: ptr.Pointer[int64](4), Domain: ptr.Pointer("bar.example.org")},
				{ID: ptr.Pointer[int64](5), Domain: ptr.Pointer("bar.example.com")},
				{ID: ptr.Pointer[int64](6), Domain: ptr.Pointer("foo.example.com")},
			},
			expected: &bunny.DNSZone{
				ID:     ptr.Pointer[int64](7),
				Domain: ptr.Pointer("foo.bar.example.com"),
			},
		},
		{
			desc:   "found apex",
			domain: "_acme-challenge.foo.bar.example.com",
			items: []*bunny.DNSZone{
				{ID: ptr.Pointer[int64](1), Domain: ptr.Pointer("example.com")},
				{ID: ptr.Pointer[int64](2), Domain: ptr.Pointer("example.org")},
				{ID: ptr.Pointer[int64](4), Domain: ptr.Pointer("bar.example.org")},
				{ID: ptr.Pointer[int64](6), Domain: ptr.Pointer("foo.example.com")},
			},
			expected: &bunny.DNSZone{
				ID:     ptr.Pointer[int64](1),
				Domain: ptr.Pointer("example.com"),
			},
		},
		{
			desc:   "not found",
			domain: "_acme-challenge.foo.bar.example.com",
			items: []*bunny.DNSZone{
				{ID: ptr.Pointer[int64](2), Domain: ptr.Pointer("example.org")},
				{ID: ptr.Pointer[int64](4), Domain: ptr.Pointer("bar.example.org")},
				{ID: ptr.Pointer[int64](6), Domain: ptr.Pointer("foo.example.com")},
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

func Test_possibleDomains(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		expected []string
	}{
		{
			desc:     "apex",
			domain:   "example.com",
			expected: []string{"example.com"},
		},
		{
			desc:     "CCTLD",
			domain:   "example.co.uk",
			expected: []string{"example.co.uk"},
		},
		{
			desc:     "long domain",
			domain:   "_acme-challenge.foo.bar.example.com",
			expected: []string{"_acme-challenge.foo.bar.example.com", "foo.bar.example.com", "bar.example.com", "example.com"},
		},
		{
			desc:   "empty",
			domain: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domains := possibleDomains(test.domain)

			assert.Equal(t, test.expected, domains)
		})
	}
}
