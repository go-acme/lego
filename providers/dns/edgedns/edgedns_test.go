package edgedns

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "TEST_DOMAIN"

var envTest = tester.NewEnvTest(
	EnvHost,
	EnvClientToken,
	EnvClientSecret,
	EnvAccessToken).
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
				EnvHost:         "A",
				EnvClientToken:  "B",
				EnvClientSecret: "C",
				EnvAccessToken:  "D",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvHost:         "",
				EnvClientToken:  "",
				EnvClientSecret: "",
				EnvAccessToken:  "",
			},
			expected: "edgedns: some credentials information are missing: AKAMAI_HOST,AKAMAI_CLIENT_TOKEN,AKAMAI_CLIENT_SECRET,AKAMAI_ACCESS_TOKEN",
		},
		{
			desc: "missing host",
			envVars: map[string]string{
				EnvHost:         "",
				EnvClientToken:  "B",
				EnvClientSecret: "C",
				EnvAccessToken:  "D",
			},
			expected: "edgedns: some credentials information are missing: AKAMAI_HOST",
		},
		{
			desc: "missing client token",
			envVars: map[string]string{
				EnvHost:         "A",
				EnvClientToken:  "",
				EnvClientSecret: "C",
				EnvAccessToken:  "D",
			},
			expected: "edgedns: some credentials information are missing: AKAMAI_CLIENT_TOKEN",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				EnvHost:         "A",
				EnvClientToken:  "B",
				EnvClientSecret: "",
				EnvAccessToken:  "D",
			},
			expected: "edgedns: some credentials information are missing: AKAMAI_CLIENT_SECRET",
		},
		{
			desc: "missing access token",
			envVars: map[string]string{
				EnvHost:         "A",
				EnvClientToken:  "B",
				EnvClientSecret: "C",
				EnvAccessToken:  "",
			},
			expected: "edgedns: some credentials information are missing: AKAMAI_ACCESS_TOKEN",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc         string
		host         string
		clientToken  string
		clientSecret string
		accessToken  string
		expected     string
	}{
		{
			desc:         "success",
			host:         "A",
			clientToken:  "B",
			clientSecret: "C",
			accessToken:  "D",
		},
		{
			desc:     "missing credentials",
			expected: "edgedns: credentials are missing",
		},
		{
			desc:         "missing host",
			host:         "",
			clientToken:  "B",
			clientSecret: "C",
			accessToken:  "D",
			expected:     "edgedns: credentials are missing",
		},
		{
			desc:         "missing client token",
			host:         "A",
			clientToken:  "",
			clientSecret: "C",
			accessToken:  "D",
			expected:     "edgedns: credentials are missing",
		},
		{
			desc:         "missing client secret",
			host:         "A",
			clientToken:  "B",
			clientSecret: "",
			accessToken:  "B",
			expected:     "edgedns: credentials are missing",
		},
		{
			desc:         "missing access token",
			host:         "A",
			clientToken:  "B",
			clientSecret: "C",
			accessToken:  "",
			expected:     "edgedns: credentials are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.ClientToken = test.clientToken
			config.ClientSecret = test.clientSecret
			config.Host = test.host
			config.AccessToken = test.accessToken

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

func TestDNSProvider_findZone(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		expected string
	}{
		{
			desc:     "Extract root record name",
			domain:   "bar.com",
			expected: "bar.com",
		},
		{
			desc:     "Extract sub record name",
			domain:   "foo.bar.com",
			expected: "bar.com",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, err := findZone(test.domain)
			require.NoError(t, err)
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

	// Present Twice to handle create / update
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

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
