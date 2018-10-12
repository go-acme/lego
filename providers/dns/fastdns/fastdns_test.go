package fastdns

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	liveTest            bool
	envTestHost         string
	envTestClientToken  string
	envTestClientSecret string
	envTestAccessToken  string
	envTestDomain       string
)

func init() {
	envTestHost = os.Getenv("AKAMAI_HOST")
	envTestClientToken = os.Getenv("AKAMAI_CLIENT_TOKEN")
	envTestClientSecret = os.Getenv("AKAMAI_CLIENT_SECRET")
	envTestAccessToken = os.Getenv("AKAMAI_ACCESS_TOKEN")
	envTestDomain = os.Getenv("AKAMAI_TEST_DOMAIN")

	if len(envTestHost) > 0 && len(envTestClientToken) > 0 && len(envTestClientSecret) > 0 && len(envTestAccessToken) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("AKAMAI_HOST", envTestHost)
	os.Setenv("AKAMAI_CLIENT_TOKEN", envTestClientToken)
	os.Setenv("AKAMAI_CLIENT_SECRET", envTestClientSecret)
	os.Setenv("AKAMAI_ACCESS_TOKEN", envTestAccessToken)
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
				"AKAMAI_HOST":          "A",
				"AKAMAI_CLIENT_TOKEN":  "B",
				"AKAMAI_CLIENT_SECRET": "C",
				"AKAMAI_ACCESS_TOKEN":  "D",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"AKAMAI_HOST":          "",
				"AKAMAI_CLIENT_TOKEN":  "",
				"AKAMAI_CLIENT_SECRET": "",
				"AKAMAI_ACCESS_TOKEN":  "",
			},
			expected: "fastdns: some credentials information are missing: AKAMAI_HOST,AKAMAI_CLIENT_TOKEN,AKAMAI_CLIENT_SECRET,AKAMAI_ACCESS_TOKEN",
		},
		{
			desc: "missing host",
			envVars: map[string]string{
				"AKAMAI_HOST":          "",
				"AKAMAI_CLIENT_TOKEN":  "B",
				"AKAMAI_CLIENT_SECRET": "C",
				"AKAMAI_ACCESS_TOKEN":  "D",
			},
			expected: "fastdns: some credentials information are missing: AKAMAI_HOST",
		},
		{
			desc: "missing client token",
			envVars: map[string]string{
				"AKAMAI_HOST":          "A",
				"AKAMAI_CLIENT_TOKEN":  "",
				"AKAMAI_CLIENT_SECRET": "C",
				"AKAMAI_ACCESS_TOKEN":  "D",
			},
			expected: "fastdns: some credentials information are missing: AKAMAI_CLIENT_TOKEN",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				"AKAMAI_HOST":          "A",
				"AKAMAI_CLIENT_TOKEN":  "B",
				"AKAMAI_CLIENT_SECRET": "",
				"AKAMAI_ACCESS_TOKEN":  "D",
			},
			expected: "fastdns: some credentials information are missing: AKAMAI_CLIENT_SECRET",
		},
		{
			desc: "missing access token",
			envVars: map[string]string{
				"AKAMAI_HOST":          "A",
				"AKAMAI_CLIENT_TOKEN":  "B",
				"AKAMAI_CLIENT_SECRET": "C",
				"AKAMAI_ACCESS_TOKEN":  "",
			},
			expected: "fastdns: some credentials information are missing: AKAMAI_ACCESS_TOKEN",
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
			expected: "fastdns: credentials are missing",
		},
		{
			desc:         "missing host",
			host:         "",
			clientToken:  "B",
			clientSecret: "C",
			accessToken:  "D",
			expected:     "fastdns: credentials are missing",
		},
		{
			desc:         "missing client token",
			host:         "A",
			clientToken:  "",
			clientSecret: "C",
			accessToken:  "D",
			expected:     "fastdns: credentials are missing",
		},
		{
			desc:         "missing client secret",
			host:         "A",
			clientToken:  "B",
			clientSecret: "",
			accessToken:  "B",
			expected:     "fastdns: credentials are missing",
		},
		{
			desc:         "missing access token",
			host:         "A",
			clientToken:  "B",
			clientSecret: "C",
			accessToken:  "",
			expected:     "fastdns: credentials are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("AKAMAI_HOST")
			os.Unsetenv("AKAMAI_CLIENT_TOKEN")
			os.Unsetenv("AKAMAI_CLIENT_SECRET")
			os.Unsetenv("AKAMAI_ACCESS_TOKEN")

			config := NewDefaultConfig()
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

func TestDNSProvider_findZoneAndRecordName(t *testing.T) {
	config := NewDefaultConfig()
	config.Host = "somehost"
	config.ClientToken = "someclienttoken"
	config.ClientSecret = "someclientsecret"
	config.AccessToken = "someaccesstoken"

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

			zone, recordName, err := provider.findZoneAndRecordName(test.fqdn, test.domain)
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

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.Host = envTestHost
	config.ClientToken = envTestClientToken
	config.ClientSecret = envTestClientSecret
	config.AccessToken = envTestAccessToken

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
