package namesilo

import (
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"NAMESILO_TTL",
	"NAMESILO_API_KEY").
	WithDomain("NAMESILO_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"NAMESILO_API_KEY": "A",
			},
		},
		{
			desc:     "missing API key",
			envVars:  map[string]string{},
			expected: "namesilo: some credentials information are missing: NAMESILO_API_KEY",
		},
		{
			desc: "unsupported TTL",
			envVars: map[string]string{
				"NAMESILO_API_KEY": "A",
				"NAMESILO_TTL":     "180",
			},
			expected: "namesilo: TTL should be in [3600, 2592000]",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
			apiKey: "A",
			ttl:    defaultTTL,
		},
		{
			desc:     "missing API key",
			ttl:      defaultTTL,
			expected: "namesilo: credentials missing: API key",
		},
		{
			desc:     "unavailable TTL",
			apiKey:   "A",
			ttl:      100,
			expected: "namesilo: TTL should be in [3600, 2592000]",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.TTL = test.ttl

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
