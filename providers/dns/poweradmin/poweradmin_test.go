package poweradmin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/poweradmin/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvBaseURL, EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvBaseURL: "https://example.com",
				EnvAPIKey:  "secret",
			},
		},
		{
			desc: "missing base URL",
			envVars: map[string]string{
				EnvBaseURL: "",
				EnvAPIKey:  "secret",
			},
			expected: "poweradmin: some credentials information are missing: POWERADMIN_BASE_URL",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvBaseURL: "https://example.com",
				EnvAPIKey:  "",
			},
			expected: "poweradmin: some credentials information are missing: POWERADMIN_API_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "poweradmin: some credentials information are missing: POWERADMIN_BASE_URL,POWERADMIN_API_KEY",
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
		baseURL  string
		apiKey   string
		expected string
	}{
		{
			desc:    "success",
			baseURL: "https://example.com",
			apiKey:  "secret",
		},
		{
			desc:     "missing base URL",
			apiKey:   "secret",
			expected: "poweradmin: missing base URL",
		},
		{
			desc:     "missing API key",
			baseURL:  "https://example.com",
			expected: "poweradmin: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "poweradmin: missing base URL",
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
			With(internal.AuthenticationHeader, "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/v2/zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("page", "1").
				With("per_page", "100"),
		).
		Route("POST /api/v2/zones/1/records",
			servermock.ResponseFromInternal("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /api/v2/zones/1/records/456",
			servermock.ResponseFromInternal("delete_record.json").
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	provider.zoneIDs["abc"] = 1
	provider.recordIDs["abc"] = 456

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
