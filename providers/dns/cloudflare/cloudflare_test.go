package cloudflare

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvEmail,
	EnvAPIKey,
	EnvDNSAPIToken,
	EnvZoneAPIToken,
	EnvBaseURL,
	altEnvEmail,
	altEnvName(EnvAPIKey),
	altEnvName(EnvDNSAPIToken),
	altEnvName(EnvZoneAPIToken)).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success email, API key",
			envVars: map[string]string{
				EnvEmail:  "test@example.com",
				EnvAPIKey: "123",
			},
		},
		{
			desc: "success API token",
			envVars: map[string]string{
				EnvDNSAPIToken: "012345abcdef",
			},
		},
		{
			desc: "success separate API tokens",
			envVars: map[string]string{
				EnvDNSAPIToken:  "012345abcdef",
				EnvZoneAPIToken: "abcdef012345",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvEmail:       "",
				EnvAPIKey:      "",
				EnvDNSAPIToken: "",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL,CLOUDFLARE_API_KEY or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN,CLOUDFLARE_ZONE_API_TOKEN",
		},
		{
			desc: "missing email",
			envVars: map[string]string{
				EnvEmail:  "",
				EnvAPIKey: "key",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN,CLOUDFLARE_ZONE_API_TOKEN",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvEmail:  "awesome@possum.com",
				EnvAPIKey: "",
			},
			expected: "cloudflare: some credentials information are missing: CLOUDFLARE_API_KEY or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN,CLOUDFLARE_ZONE_API_TOKEN",
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
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderWithToken(t *testing.T) {
	type expected struct {
		dnsToken   string
		zoneToken  string
		sameClient bool
		error      string
	}

	testCases := []struct {
		desc string

		// test input
		envVars map[string]string

		// expectations
		expected expected
	}{
		{
			desc: "same client when zone token is missing",
			envVars: map[string]string{
				EnvDNSAPIToken: "123",
			},
			expected: expected{
				dnsToken:   "123",
				zoneToken:  "123",
				sameClient: true,
			},
		},
		{
			desc: "same client when zone token equals dns token",
			envVars: map[string]string{
				EnvDNSAPIToken:  "123",
				EnvZoneAPIToken: "123",
			},
			expected: expected{
				dnsToken:   "123",
				zoneToken:  "123",
				sameClient: true,
			},
		},
		{
			desc: "failure when only zone api given",
			envVars: map[string]string{
				EnvZoneAPIToken: "123",
			},
			expected: expected{
				error: "cloudflare: some credentials information are missing: CLOUDFLARE_EMAIL,CLOUDFLARE_API_KEY or some credentials information are missing: CLOUDFLARE_DNS_API_TOKEN",
			},
		},
		{
			desc: "different clients when zone and dns token differ",
			envVars: map[string]string{
				EnvDNSAPIToken:  "123",
				EnvZoneAPIToken: "abc",
			},
			expected: expected{
				dnsToken:   "123",
				zoneToken:  "abc",
				sameClient: false,
			},
		},
		{
			desc: "aliases work as expected", // CLOUDFLARE_* takes precedence over CF_*
			envVars: map[string]string{
				EnvDNSAPIToken:              "123",
				altEnvName(EnvDNSAPIToken):  "456",
				EnvZoneAPIToken:             "abc",
				altEnvName(EnvZoneAPIToken): "def",
			},
			expected: expected{
				dnsToken:   "123",
				zoneToken:  "abc",
				sameClient: false,
			},
		},
	}

	defer envTest.RestoreEnv()

	localEnvTest := tester.NewEnvTest(
		EnvDNSAPIToken, altEnvName(EnvDNSAPIToken),
		EnvZoneAPIToken, altEnvName(EnvZoneAPIToken),
	).WithDomain(envDomain)

	envTest.ClearEnv()

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer localEnvTest.RestoreEnv()

			localEnvTest.ClearEnv()
			localEnvTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected.error != "" {
				require.EqualError(t, err, test.expected.error)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, p)
			assert.Equal(t, test.expected.dnsToken, p.config.AuthToken)
			assert.Equal(t, test.expected.zoneToken, p.config.ZoneToken)

			if test.expected.sameClient {
				assert.Equal(t, p.client.clientRead, p.client.clientEdit)
			} else {
				assert.NotEqual(t, p.client.clientRead, p.client.clientEdit)
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
			expected: "cloudflare: invalid credentials: authEmail, authKey or authToken must be set",
		},
		{
			desc:     "missing email",
			authKey:  "123",
			expected: "cloudflare: invalid credentials: authEmail and authKey must be set together",
		},
		{
			desc:      "missing api key",
			authEmail: "test@example.com",
			expected:  "cloudflare: invalid credentials: authEmail and authKey must be set together",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthEmail = test.authEmail
			config.AuthKey = test.authKey
			config.AuthToken = test.authToken

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
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

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.AuthEmail = "foo@example.com"
			config.AuthKey = "secret"
			config.BaseURL = server.URL
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithRegexp("User-Agent", `goacme-lego/[0-9.]+ \(.+\)`).
			With("X-Auth-Email", "foo@example.com").
			With("X-Auth-Key", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		// https://developers.cloudflare.com/api/resources/zones/methods/list/
		Route("GET /zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com").
				With("per_page", "50")).
		// https://developers.cloudflare.com/api/resources/dns/subresources/records/methods/create/
		Route("POST /zones/023e105f4ecef8ad9ca31a8372d0c353/dns_records",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckHeader().
				WithContentType("application/json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json")).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		// https://developers.cloudflare.com/api/resources/zones/methods/list/
		Route("GET /zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com").
				With("per_page", "50")).
		// https://developers.cloudflare.com/api/resources/dns/subresources/records/methods/delete/
		Route("DELETE /zones/023e105f4ecef8ad9ca31a8372d0c353/dns_records/xxx",
			servermock.ResponseFromInternal("delete_record.json")).
		Build(t)

	token := "abc"

	provider.recordIDsMu.Lock()
	provider.recordIDs["abc"] = "xxx"
	provider.recordIDsMu.Unlock()

	err := provider.CleanUp("example.com", token, "123d==")
	require.NoError(t, err)
}
