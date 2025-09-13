package keyhelp

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/keyhelp/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvBaseURL, EnvAPIKey).
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
				EnvBaseURL: "https://keyhelp.example.com",
				EnvAPIKey:  "secret",
			},
		},
		{
			desc: "missing base URL",
			envVars: map[string]string{
				EnvAPIKey: "secret",
			},
			expected: "keyhelp: some credentials information are missing: KEYHELP_BASE_URL",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvBaseURL: "https://keyhelp.example.com",
			},
			expected: "keyhelp: some credentials information are missing: KEYHELP_API_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "keyhelp: some credentials information are missing: KEYHELP_BASE_URL,KEYHELP_API_KEY",
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
			baseURL: "https://keyhelp.example.com",
			apiKey:  "secret",
		},
		{
			desc:     "missing base URL",
			apiKey:   "secret",
			expected: "keyhelp: missing base URL",
		},
		{
			desc:     "missing API key",
			baseURL:  "https://keyhelp.example.com",
			expected: "keyhelp: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "keyhelp: missing base URL",
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		config := NewDefaultConfig()
		config.HTTPClient = server.Client()
		config.APIKey = "secret"
		config.BaseURL = server.URL

		return NewDNSProviderConfig(config)
	},
		servermock.CheckHeader().
			With(internal.APIKeyHeader, "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/v2/domains",
			servermock.ResponseFromInternal("get_domains.json"),
			servermock.CheckQueryParameter().
				With("sort", "domain_utf8").
				Strict()).
		Route("GET /api/v2/dns/8",
			servermock.ResponseFromInternal("get_domain_records.json")).
		Route("PUT /api/v2/dns/8",
			servermock.ResponseFromInternal("update_domain_records.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_domain_records-request.json")).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, 8, provider.domainIDs["abc"])
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /api/v2/dns/8",
			servermock.ResponseFromInternal("get_domain_records2.json")).
		Route("PUT /api/v2/dns/8",
			servermock.ResponseFromInternal("update_domain_records.json"),
			servermock.CheckRequestJSONBodyFromInternal("update_domain_records-request2.json")).
		Build(t)

	provider.domainIDs["abc"] = 8

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
