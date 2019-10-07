package cloudflare

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"CLOUDFLARE_EMAIL",
	"CLOUDFLARE_API_KEY",
	"CLOUDFLARE_DNS_API_TOKEN",
	"CLOUDFLARE_ZONE_API_TOKEN").
	WithDomain("CLOUDFLARE_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success email, API key",
			envVars: map[string]string{
				"CLOUDFLARE_EMAIL":   "test@example.com",
				"CLOUDFLARE_API_KEY": "123",
			},
		},
		{
			desc: "success API token",
			envVars: map[string]string{
				"CLOUDFLARE_DNS_API_TOKEN": "012345abcdef",
			},
		},
		{
			desc: "success separate API tokens",
			envVars: map[string]string{
				"CLOUDFLARE_DNS_API_TOKEN":  "012345abcdef",
				"CLOUDFLARE_ZONE_API_TOKEN": "abcdef012345",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"CLOUDFLARE_EMAIL":         "",
				"CLOUDFLARE_API_KEY":       "",
				"CLOUDFLARE_DNS_API_TOKEN": "",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL,CLOUDFLARE_API_KEY or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				"CLOUDFLARE_EMAIL":   "",
				"CLOUDFLARE_API_KEY": "key",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				"CLOUDFLARE_EMAIL":   "awesome@possum.com",
				"CLOUDFLARE_API_KEY": "",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_API_KEY or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN",
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
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.dns)
				assert.NotNil(t, p.zone)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderWithToken(t *testing.T) {
	testCases := []struct {
		desc string

		// test input
		envVars map[string]string

		// expectations
		dnsToken    string
		zoneToken   string
		sameClient  bool
		expectError string
	}{
		{
			desc: "same client when zone token is missing",
			envVars: map[string]string{
				"CLOUDFLARE_DNS_API_TOKEN": "123",
			},
			dnsToken:   "123",
			zoneToken:  "",
			sameClient: true,
		},
		{
			desc: "same client when zone token equals dns token",
			envVars: map[string]string{
				"CLOUDFLARE_DNS_API_TOKEN":  "123",
				"CLOUDFLARE_ZONE_API_TOKEN": "123",
			},
			dnsToken:   "123",
			zoneToken:  "123",
			sameClient: true,
		},
		{
			desc: "failure when only zone api given",
			envVars: map[string]string{
				"CLOUDFLARE_ZONE_API_TOKEN": "123",
			},
			expectError: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL,CLOUDFLARE_API_KEY or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN",
		},
		{
			desc: "different clients when zone and dns token differ",
			envVars: map[string]string{
				"CLOUDFLARE_DNS_API_TOKEN":  "123",
				"CLOUDFLARE_ZONE_API_TOKEN": "abc",
			},
			dnsToken:   "123",
			zoneToken:  "abc",
			sameClient: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expectError != "" {
				require.EqualError(t, err, test.expectError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, p)
			assert.Equal(t, test.dnsToken, p.config.AuthToken)
			assert.Equal(t, test.zoneToken, p.config.ZoneToken)
			if test.sameClient {
				assert.Equal(t, p.zone, p.dns)
			} else {
				assert.NotEqual(t, p.zone, p.dns)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		authEmail string
		authKey   string
		authToken string
		expected  string
	}{
		{
			desc:      "success with email and api key",
			authEmail: "test@example.com",
			authKey:   "123",
		},
		{
			desc:      "success with api token",
			authToken: "012345abcdef",
		},
		{
			desc:      "prefer api token",
			authToken: "012345abcdef",
			authEmail: "test@example.com",
			authKey:   "123",
		},
		{
			desc:     "missing credentials",
			expected: "invalid credentials: key & email must not be empty",
		},
		{
			desc:     "missing email",
			authKey:  "123",
			expected: "invalid credentials: key & email must not be empty",
		},
		{
			desc:      "missing api key",
			authEmail: "test@example.com",
			expected:  "invalid credentials: key & email must not be empty",
		},
		{
			desc:      "missing api token, fallback to api key/email",
			authToken: "",
			expected:  "invalid credentials: key & email must not be empty",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthEmail = test.authEmail
			config.AuthKey = test.authKey
			config.AuthToken = test.authToken

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.dns)
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
