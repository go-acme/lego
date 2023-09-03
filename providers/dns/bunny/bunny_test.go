package bunny

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
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

func Test_splitDomain(t *testing.T) {
	type expected struct {
		root       string
		sub        string
		requireErr require.ErrorAssertionFunc
	}

	testCases := []struct {
		desc     string
		domain   string
		expected expected
	}{
		{
			desc:   "empty",
			domain: "",
			expected: expected{
				requireErr: require.Error,
			},
		},
		{
			desc:   "2 levels",
			domain: "example.com",
			expected: expected{
				root:       "example.com",
				sub:        "",
				requireErr: require.NoError,
			},
		},
		{
			desc:   "3 levels",
			domain: "_acme-challenge.example.com",
			expected: expected{
				root:       "example.com",
				sub:        "_acme-challenge",
				requireErr: require.NoError,
			},
		},
		{
			desc:   "4 levels",
			domain: "_acme-challenge.sub.example.com",
			expected: expected{
				root:       "example.com",
				sub:        "_acme-challenge.sub",
				requireErr: require.NoError,
			},
		},
		{
			desc:   "5 levels",
			domain: "_acme-challenge.my.sub.example.com",
			expected: expected{
				root:       "example.com",
				sub:        "_acme-challenge.my.sub",
				requireErr: require.NoError,
			},
		},
		{
			desc:   "6 levels",
			domain: "_acme-challenge.my.sub.sub.example.com",
			expected: expected{
				root:       "example.com",
				sub:        "_acme-challenge.my.sub.sub",
				requireErr: require.NoError,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			root, sub, err := splitDomain(test.domain)
			test.expected.requireErr(t, err)

			assert.Equal(t, test.expected.root, root)
			assert.Equal(t, test.expected.sub, sub)
		})
	}
}
