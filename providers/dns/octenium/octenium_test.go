package octenium

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "secret",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "octenium: some credentials information are missing: OCTENIUM_API_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "octenium: some credentials information are missing: OCTENIUM_API_KEY",
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
		expected string
	}{
		{
			desc:   "success",
			apiKey: "secret",
		},
		{
			desc:     "missing API key",
			expected: "octenium: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "octenium: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
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
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
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
			WithAccept("application/json").
			With("X-Api-Key", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains",
			servermock.ResponseFromFixture("list_domains.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain-name", "example.com")).
		Route("POST /domains/dns-records/add",
			servermock.ResponseFromFixture("add_dns_record.json"),
			servermock.CheckHeader().
				WithContentType("application/x-www-form-urlencoded"),
			servermock.CheckForm().Strict().
				With("order-id", "2976").
				With("name", "_acme-challenge.example.com.").
				With("ttl", "120").
				With("type", "TXT").
				With("value", "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI")).
		Build(t)

	err := provider.Present("example.com", "", "foobar")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /domains/dns-records/list",
			servermock.ResponseFromFixture("list_dns_records.json"),
			servermock.CheckHeader().
				WithContentType("application/x-www-form-urlencoded"),
			servermock.CheckForm().Strict().
				With("order-id", "2976").
				With("types[]", "TXT")).
		Route("POST /domains/dns-records/delete",
			servermock.ResponseFromFixture("delete_dns_record.json"),
			servermock.CheckHeader().
				WithContentType("application/x-www-form-urlencoded"),
			servermock.CheckForm().Strict().
				With("order-id", "2976").
				With("line", "123")).
		Build(t)

	provider.domainIDs["token"] = "2976"

	err := provider.CleanUp("example.com", "token", "foobar")
	require.NoError(t, err)
}
