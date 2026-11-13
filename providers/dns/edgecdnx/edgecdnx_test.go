package edgecdnx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/edgecdnx/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvProjectID).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey:    "secret",
				EnvProjectID: "p123",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIKey:    "",
				EnvProjectID: "p123",
			},
			expected: "edgecdnx: some credentials information are missing: EDGECDNX_API_KEY",
		},
		{
			desc: "missing project ID",
			envVars: map[string]string{
				EnvAPIKey:    "secret",
				EnvProjectID: "",
			},
			expected: "edgecdnx: some credentials information are missing: EDGECDNX_PROJECT_ID",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "edgecdnx: some credentials information are missing: EDGECDNX_API_KEY,EDGECDNX_PROJECT_ID",
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
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		apiKey    string
		projectID string
		expected  string
	}{
		{
			desc:      "success",
			apiKey:    "secret",
			projectID: "p123",
		},
		{
			desc:      "missing API key",
			projectID: "p123",
			expected:  "edgecdnx: credentials missing",
		},
		{
			desc:     "missing project ID",
			apiKey:   "secret",
			expected: "edgecdnx: missing project ID",
		},
		{
			desc:     "missing credentials",
			expected: "edgecdnx: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.ProjectID = test.projectID

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
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

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIKey = "secret"
			config.ProjectID = "p123"
			config.APIURL = server.URL
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("ApiKey secret"),
	)
}

func TestDNSProvider_Present_update(t *testing.T) {
	provider := mockBuilder().
		Route("GET /project/p123/zones",
			servermock.ResponseFromInternal("zones.json"),
		).
		Route("GET /project/p123/zones/z123/dns-endpoints",
			servermock.ResponseFromInternal("dns_endpoints.json"),
		).
		Route("PATCH /project/p123/zones/z123/dns-endpoints/r123",
			servermock.ResponseFromInternal("update_dns_endpoint.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_dns_endpoint-request.json"),
		).
		Route("/",
			servermock.DumpRequest(),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"abc": "z123"}, provider.zoneIDs)
}

func TestDNSProvider_Present_create(t *testing.T) {
	provider := mockBuilder().
		Route("GET /project/p123/zones",
			servermock.ResponseFromInternal("zones.json"),
		).
		Route("GET /project/p123/zones/z123/dns-endpoints",
			servermock.RawStringResponse(`[]`),
		).
		Route("POST /project/p123/zones/z123/dns-endpoints",
			servermock.ResponseFromInternal("create_dns_endpoint.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_dns_endpoint-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"abc": "z123"}, provider.zoneIDs)
}

func TestDNSProvider_CleanUp_update(t *testing.T) {
	provider := mockBuilder().
		Route("GET /project/p123/zones",
			servermock.ResponseFromInternal("zones.json"),
		).
		Route("GET /project/p123/zones/z123/dns-endpoints",
			servermock.ResponseFromInternal("dns_endpoints.json"),
		).
		Route("PATCH /project/p123/zones/z123/dns-endpoints/r123",
			servermock.ResponseFromInternal("update_dns_endpoint.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_dns_endpoint-request2.json"),
		).
		Build(t)

	provider.zoneIDs["abc"] = "z123"

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.zoneIDs)
}

func TestDNSProvider_CleanUp_delete(t *testing.T) {
	provider := mockBuilder().
		Route("GET /project/p123/zones",
			servermock.ResponseFromInternal("zones.json"),
		).
		Route("GET /project/p123/zones/z123/dns-endpoints",
			servermock.ResponseFromInternal("dns_endpoints2.json"),
		).
		Route("DELETE /project/p123/zones/z123/dns-endpoints/r123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckQueryParameter().Strict(),
		).
		Build(t)

	provider.zoneIDs["abc"] = "z123"

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.zoneIDs)
}

func Test_findZoneID(t *testing.T) {
	zones := []internal.Zone{
		{Metadata: internal.Metadata{Name: "z1"}, Spec: internal.ZoneSpec{Zone: "a.example"}},
		{Metadata: internal.Metadata{Name: "z2"}, Spec: internal.ZoneSpec{Zone: "b.example"}},
		{Metadata: internal.Metadata{Name: "z3"}, Spec: internal.ZoneSpec{Zone: "c.a.b.example"}},
		{Metadata: internal.Metadata{Name: "z4"}, Spec: internal.ZoneSpec{Zone: "b.a.example"}},
		{Metadata: internal.Metadata{Name: "z5"}, Spec: internal.ZoneSpec{Zone: "c.b.a.example"}},
		{Metadata: internal.Metadata{Name: "z6"}, Spec: internal.ZoneSpec{Zone: "a.b.example"}},
	}

	testCases := []struct {
		desc     string
		fqdn     string
		expected string
	}{
		{
			desc:     "subdomain",
			fqdn:     "d.a.b.example.",
			expected: "z6",
		},
		{
			desc:     "exact zone",
			fqdn:     "a.example.",
			expected: "z1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			id, err := findZoneID(zones, test.fqdn)
			require.NoError(t, err)

			assert.Equal(t, test.expected, id)
		})
	}
}

func Test_findDNSEndpoint(t *testing.T) {
	records := []internal.Record{
		{
			Spec: internal.RecordSpec{
				DNSName:       "example.org",
				Type:          "A",
				RoutingPolicy: internal.RoutingPolicySimple,
			},
		},
		{
			Spec: internal.RecordSpec{
				DNSName:       "example.com",
				Type:          "TXT",
				RoutingPolicy: "Weighted",
			},
		},
		{
			Spec: internal.RecordSpec{
				DNSName:       "example.org",
				Type:          "TXT",
				RoutingPolicy: internal.RoutingPolicySimple,
			},
		},
		{
			Spec: internal.RecordSpec{
				DNSName:       "example.com",
				Type:          "TXT",
				RoutingPolicy: internal.RoutingPolicySimple,
			},
		},
	}

	endpoint, err := findDNSEndpoint(records, "example.com")
	require.NoError(t, err)

	expected := internal.Record{
		Spec: internal.RecordSpec{
			DNSName:       "example.com",
			Type:          "TXT",
			RoutingPolicy: internal.RoutingPolicySimple,
		},
	}

	assert.Equal(t, expected, endpoint)
}
