package anxcloud

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.anx.io/go-anxcloud/pkg/clouddns/zone"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvToken,
	EnvAPIURL).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success with token",
			envVars: map[string]string{
				EnvToken: "test-token-123",
			},
		},
		{
			desc: "success with token and custom API URL",
			envVars: map[string]string{
				EnvToken:  "test-token-123",
				EnvAPIURL: "https://custom.api.example.com",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvToken: "",
			},
			expected: "anxcloud: some credentials information are missing: ANEXIA_TOKEN",
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
				assert.NotNil(t, p.api)
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
		apiURL   string
		expected string
	}{
		{
			desc:  "success with token",
			token: "test-token-123",
		},
		{
			desc:   "success with token and custom API URL",
			token:  "test-token-123",
			apiURL: "https://custom.api.example.com",
		},
		{
			desc:     "missing token",
			token:    "",
			expected: "anxcloud: incomplete credentials, missing token",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Token = test.token
			config.APIURL = test.apiURL

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.api)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig_Nil(t *testing.T) {
	p, err := NewDNSProviderConfig(nil)

	require.EqualError(t, err, "anxcloud: the configuration of the DNS provider is nil")
	require.Nil(t, p)
}

func TestDNSProvider_Timeout(t *testing.T) {
	config := NewDefaultConfig()
	config.Token = "test-token"
	config.PropagationTimeout = 5 * time.Minute
	config.PollingInterval = 30 * time.Second

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	timeout, interval := p.Timeout()
	assert.Equal(t, 5*time.Minute, timeout)
	assert.Equal(t, 30*time.Second, interval)
}

func TestExtractRecordName(t *testing.T) {
	testCases := []struct {
		desc     string
		fqdn     string
		zone     string
		expected string
	}{
		{
			desc:     "root domain",
			fqdn:     "example.com.",
			zone:     "example.com.",
			expected: "@",
		},
		{
			desc:     "subdomain",
			fqdn:     "_acme-challenge.example.com.",
			zone:     "example.com.",
			expected: "_acme-challenge",
		},
		{
			desc:     "nested subdomain",
			fqdn:     "_acme-challenge.sub.example.com.",
			zone:     "example.com.",
			expected: "_acme-challenge.sub",
		},
		{
			desc:     "zone with subdomain",
			fqdn:     "_acme-challenge.test.sub.example.com.",
			zone:     "sub.example.com.",
			expected: "_acme-challenge.test",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			result := extractRecordName(test.fqdn, test.zone)
			assert.Equal(t, test.expected, result)
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

func mustParseUUID(s string) uuid.UUID {
	u, err := uuid.FromString(s)
	if err != nil {
		panic(err)
	}
	return u
}

func createZoneResponse(rdata string) zone.Zone {
	ttl := 300
	return zone.Zone{
		Definition: &zone.Definition{
			ZoneName:   "example.com",
			Name:       "example.com",
			IsMaster:   true,
			DNSSecMode: "managed",
			AdminEmail: "admin@example.com",
			Refresh:    10800,
			Retry:      3600,
			Expire:     604800,
			TTL:        86400,
		},
		Customer:        "ANX12345",
		IsEditable:      true,
		ValidationLevel: 0,
		DeploymentLevel: 0,
		Revisions: []zone.Revision{
			{
				Identifier: mustParseUUID("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
				Serial:     1,
				State:      "deployed",
				Records: []zone.Record{
					{
						Identifier: mustParseUUID("12345678-1234-1234-1234-123456789abc"),
						Immutable:  false,
						Name:       "_acme-challenge",
						RData:      rdata,
						Region:     "",
						TTL:        &ttl,
						Type:       "TXT",
					},
				},
			},
		},
	}
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Token = "test-token-123"
			config.APIURL = server.URL
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithRegexp("User-Agent", `go-anxcloud/.+ \(.+\)`).
			With("Authorization", "Token test-token-123"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	// Compute the expected challenge value
	info := dns01.GetChallengeInfo("example.com", "123d==")

	// Expected request body
	expectedRequest := zone.RecordRequest{
		Name:   "_acme-challenge",
		Type:   "TXT",
		RData:  info.Value,
		Region: "",
		TTL:    300,
	}

	provider := mockBuilder().
		// POST /api/clouddns/v1/zone.json/{zoneName}/records
		Route("POST /api/clouddns/v1/zone.json/{zoneName}/records",
			servermock.JSONEncode(createZoneResponse(info.Value)),
			servermock.CheckHeader().
				WithContentType("application/json; charset=utf-8"),
			servermock.CheckRequestJSONBodyFromStruct(expectedRequest)).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		// DELETE /api/clouddns/v1/zone.json/{zoneName}/records/{recordUUID}
		Route("DELETE /api/clouddns/v1/zone.json/{zoneName}/records/{recordUUID}",
			servermock.Noop()).
		Build(t)

	token := "abc"

	provider.recordIDsMu.Lock()
	provider.recordIDs["abc"] = mustParseUUID("12345678-1234-1234-1234-123456789abc")
	provider.recordIDsMu.Unlock()

	err := provider.CleanUp("example.com", token, "123d==")
	require.NoError(t, err)
}
