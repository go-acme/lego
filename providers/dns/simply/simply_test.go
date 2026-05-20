package simply

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAccountName, EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAccountName: "S000000",
				EnvAPIKey:      "secret",
			},
		},
		{
			desc: "missing credentials: account name",
			envVars: map[string]string{
				EnvAccountName: "",
				EnvAPIKey:      "secret",
			},
			expected: "simply: some credentials information are missing: SIMPLY_ACCOUNT_NAME",
		},
		{
			desc: "missing credentials: api key",
			envVars: map[string]string{
				EnvAccountName: "S000000",
				EnvAPIKey:      "",
			},
			expected: "simply: some credentials information are missing: SIMPLY_API_KEY",
		},
		{
			desc: "missing credentials: all",
			envVars: map[string]string{
				EnvAccountName: "",
				EnvAPIKey:      "",
			},
			expected: "simply: some credentials information are missing: SIMPLY_ACCOUNT_NAME,SIMPLY_API_KEY",
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
		desc        string
		accountName string
		apiKey      string
		expected    string
	}{
		{
			desc:        "success",
			accountName: "S000000",
			apiKey:      "secret",
		},
		{
			desc:     "missing account name",
			apiKey:   "secret",
			expected: "simply: missing credentials: account name",
		},
		{
			desc:        "missing api key",
			accountName: "S000000",
			expected:    "simply: missing credentials: api key",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccountName = test.accountName
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
			config.AccountName = "S000000"
			config.APIKey = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /my/products/example.com/dns/records/",
			servermock.ResponseFromInternal("add_record.json"),
			servermock.CheckRequestJSONBodyFromInternal("add_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /my/products/example.com/dns/records/123456789/",
			servermock.ResponseFromInternal("success.json"),
		).
		Build(t)

	provider.recordIDs["abc"] = 123456789

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
