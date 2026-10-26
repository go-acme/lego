package cloudpress

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey, EnvBaseURL).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey:  "secret",
				EnvBaseURL: "https://example.com",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIKey:  "",
				EnvBaseURL: "https://example.com",
			},
			expected: "cloudpress: some credentials information are missing: CLOUDPRESS_API_KEY",
		},
		{
			desc: "missing API URL",
			envVars: map[string]string{
				EnvAPIKey:  "secret",
				EnvBaseURL: "",
			},
			expected: "cloudpress: some credentials information are missing: CLOUDPRESS_BASE_URL",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "cloudpress: some credentials information are missing: CLOUDPRESS_BASE_URL,CLOUDPRESS_API_KEY",
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
		desc     string
		apiKey   string
		baseURL  string
		expected string
	}{
		{
			desc:    "success",
			apiKey:  "secret",
			baseURL: "https://example.com",
		},
		{
			desc:     "missing API key",
			expected: "cloudpress: credentials missing",
			baseURL:  "https://example.com",
		},
		{
			desc:     "missing API URL",
			expected: "cloudpress: missing base URL",
			apiKey:   "secret",
		},
		{
			desc:     "missing credentials",
			expected: "cloudpress: missing base URL",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = test.baseURL
			config.APIKey = test.apiKey

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
			config.BaseURL = server.URL
			config.APIKey = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/dns_zones",
			servermock.ResponseFromInternal("list_dns_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("page", "1").
				With("per_page", "100"),
		).
		Route("POST /api/dns_zones/z123/records",
			servermock.ResponseFromInternal("create_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"abc": "r123"}, provider.recordIDs)
	assert.Equal(t, map[string]string{"abc": "z123"}, provider.zoneIDs)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /api/dns_zones/z123/records/r123",
			servermock.Noop().
				WithStatusCode(http.StatusAccepted),
		).
		Build(t)

	provider.recordIDs["abc"] = "r123"
	provider.zoneIDs["abc"] = "z123"

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.recordIDs)
	assert.Empty(t, provider.zoneIDs)
}
