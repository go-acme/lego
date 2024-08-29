package mittwald

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/providers/dns/mittwald/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvToken: "secret",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvToken: "",
			},
			expected: "mittwald: some credentials information are missing: MITTWALD_TOKEN",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				assert.NoError(t, err)
				assert.NotNil(t, p)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		token    string
		ttl      int
		expected string
	}{
		{
			desc:  "success",
			token: "secret",
		},
		{
			desc:     "missing credentials",
			expected: "mittwald: some credentials information are missing",
		},
		{
			desc:     "invalid TTL",
			token:    "secret",
			ttl:      10,
			expected: "mittwald: invalid TTL, TTL (10) must be greater than 300",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Token = test.token

			if test.ttl > 0 {
				config.TTL = test.ttl
			}

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				assert.NoError(t, err)
				assert.NotNil(t, p)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func Test_findDomain(t *testing.T) {
	domains := []internal.Domain{
		{
			Domain:    "example.com",
			ProjectID: "a1",
		},
		{
			Domain:    "foo.example.com",
			ProjectID: "a2",
		},
		{
			Domain:    "example.org",
			ProjectID: "b1",
		},
		{
			Domain:    "foo.example.org",
			ProjectID: "b2",
		},
		{
			Domain:    "test.example.org",
			ProjectID: "b3",
		},
	}

	testCases := []struct {
		desc     string
		fqdn     string
		expected internal.Domain
	}{
		{
			desc:     "exact match",
			fqdn:     "example.org.",
			expected: internal.Domain{Domain: "example.org", ProjectID: "b1"},
		},
		{
			desc:     "1 level parent",
			fqdn:     "_acme-challenge.test.example.org.",
			expected: internal.Domain{Domain: "test.example.org", ProjectID: "b3"},
		},
		{
			desc:     "2 levels parent",
			fqdn:     "_acme-challenge.test.example.com.",
			expected: internal.Domain{Domain: "example.com", ProjectID: "a1"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domain, err := findDomain(domains, test.fqdn)
			require.NoError(t, err)

			assert.Equal(t, test.expected, domain)
		})
	}
}

func Test_findZone(t *testing.T) {
	zones := []internal.DNSZone{
		{
			Domain: "example.com",
			ID:     "a1",
		},
		{
			Domain: "foo.example.com",
			ID:     "a2",
		},
		{
			Domain: "example.org",
			ID:     "b1",
		},
		{
			Domain: "foo.example.org",
			ID:     "b2",
		},
		{
			Domain: "test.example.org",
			ID:     "b3",
		},
	}

	testCases := []struct {
		desc     string
		fqdn     string
		expected internal.DNSZone
	}{
		{
			desc:     "exact match",
			fqdn:     "example.org.",
			expected: internal.DNSZone{Domain: "example.org", ID: "b1"},
		},
		{
			desc:     "1 level parent",
			fqdn:     "_acme-challenge.test.example.org.",
			expected: internal.DNSZone{Domain: "test.example.org", ID: "b3"},
		},
		{
			desc:     "2 levels parent",
			fqdn:     "_acme-challenge.test.example.com.",
			expected: internal.DNSZone{Domain: "example.com", ID: "a1"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, err := findZone(zones, test.fqdn)
			require.NoError(t, err)

			assert.Equal(t, test.expected, zone)
		})
	}
}
