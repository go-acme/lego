package pdns

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIURL,
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
				EnvAPIURL: "http://example.com",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIKey: "",
				EnvAPIURL: "",
			},
			expected: "pdns: some credentials information are missing: PDNS_API_KEY,PDNS_API_URL",
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIKey: "",
				EnvAPIURL: "http://example.com",
			},
			expected: "pdns: some credentials information are missing: PDNS_API_KEY",
		},
		{
			desc: "missing API URL",
			envVars: map[string]string{
				EnvAPIKey: "123",
				EnvAPIURL: "",
			},
			expected: "pdns: some credentials information are missing: PDNS_API_URL",
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
		desc             string
		apiKey           string
		customAPIVersion int
		host             *url.URL
		expected         string
	}{
		{
			desc:   "success",
			apiKey: "123",
			host:   mustParse("http://example.com"),
		},
		{
			desc:             "success custom API version",
			apiKey:           "123",
			customAPIVersion: 1,
			host:             mustParse("http://example.com"),
		},
		{
			desc:     "missing credentials",
			expected: "pdns: API key missing",
		},
		{
			desc:     "missing API key",
			apiKey:   "",
			host:     mustParse("http://example.com"),
			expected: "pdns: API key missing",
		},
		{
			desc:     "missing host",
			apiKey:   "123",
			expected: "pdns: API URL missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.Host = test.host
			config.APIVersion = test.customAPIVersion

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

func TestLivePresentAndCleanup(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123e==")
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123e==")
	require.NoError(t, err)
}

func mustParse(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return u
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIKey = "secret"
			config.Host, _ = url.Parse(server.URL)
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("X-Api-Key", "secret"),
	).Route("GET /api",
		servermock.JSONEncode([]map[string]any{{"url": "/x", "version": 4}}),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/v4/servers/localhost/zones/example.org.",
			servermock.ResponseFromInternal("zone.json"),
		).
		Route("PATCH /api/v4/servers/localhost/zones/example.org.",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromInternal("zone_a.json"),
		).
		Route("PUT /api/v4/servers/localhost/zones/example.org./notify",
			servermock.Noop(),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.org", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_existing_TXT(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/v4/servers/localhost/zones/example.org.",
			servermock.ResponseFromInternal("zones_a.json"),
		).
		Route("PATCH /api/v4/servers/localhost/zones/example.org.",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromInternal("zone_b.json"),
		).
		Route("PUT /api/v4/servers/localhost/zones/example.org./notify",
			servermock.Noop(),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.org", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/v4/servers/localhost/zones/example.org.",
			servermock.ResponseFromInternal("zone_b.json"),
		).
		Route("PATCH /api/v4/servers/localhost/zones",
			servermock.ResponseFromInternal("zone.json"),
			servermock.CheckRequestJSONBodyFromInternal("zone_update-request.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.org", "abc", "123d==")
	require.NoError(t, err)
}
