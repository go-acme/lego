package feno

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
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
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "feno: some credentials information are missing: FENO_API_KEY",
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
		ttl      int
		expected string
	}{
		{
			desc:   "success",
			apiKey: "secret",
			ttl:    120,
		},
		{
			desc:     "missing credentials",
			ttl:      120,
			expected: "feno: credentials missing",
		},
		{
			desc:     "invalid TTL",
			apiKey:   "secret",
			ttl:      10,
			expected: "feno: invalid TTL, TTL (10) must be greater than 15",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.TTL = test.ttl

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
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.no",
			servermock.ResponseFromInternal("get_domain.json")).
		Route("POST /domains/example.no/dns",
			servermock.ResponseFromInternal("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.no", "abc", "123d==")
	require.NoError(t, err)

	require.Equal(t, "example.no", provider.zones["abc"])
	require.Equal(t, 481522, provider.recordIDs["abc"])
}

func TestDNSProvider_Present_subdomain(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/www.example.no",
			servermock.ResponseFromInternal("error_not_managed.json").
				WithStatusCode(http.StatusNotFound)).
		Route("GET /domains/example.no",
			servermock.ResponseFromInternal("get_domain.json")).
		Route("POST /domains/example.no/dns",
			servermock.ResponseFromInternal("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBody(`{"type":"TXT","name":"_acme-challenge.www","value":"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY","ttl":120}`),
		).
		Build(t)

	err := provider.Present(t.Context(), "www.example.no", "abc", "123d==")
	require.NoError(t, err)

	require.Equal(t, "example.no", provider.zones["abc"])
}

func TestDNSProvider_Present_notFenoNameservers(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.no",
			servermock.ResponseFromInternal("get_domain_not_feno_ns.json")).
		Build(t)

	err := provider.Present(t.Context(), "example.no", "abc", "123d==")
	require.EqualError(t, err, `feno: could not find zone for domain "example.no": the zone "example.no" does not use FENO nameservers`)
}

func TestDNSProvider_Present_zoneNotFound(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/{domain}",
			servermock.ResponseFromInternal("error_not_managed.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	err := provider.Present(t.Context(), "www.example.no", "abc", "123d==")
	require.EqualError(t, err, `feno: could not find zone for domain "www.example.no": no zone found for "_acme-challenge.www.example.no.": 404: Domain not found or not managed by this account (DOMAIN_NOT_MANAGED)`)
}

func TestDNSProvider_Present_insufficientScope(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/www.example.no",
			servermock.ResponseFromInternal("error_insufficient_scope.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	err := provider.Present(t.Context(), "www.example.no", "abc", "123d==")
	require.EqualError(t, err, `feno: could not find zone for domain "www.example.no": 403: This key does not hold the required scope (INSUFFICIENT_SCOPE)`)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /domains/example.no/dns/481522",
			servermock.ResponseFromInternal("delete_record.json")).
		Build(t)

	token := "abc"

	provider.zones[token] = "example.no"
	provider.recordIDs[token] = 481522

	err := provider.CleanUp(t.Context(), "example.no", token, "123d==")
	require.NoError(t, err)

	require.Empty(t, provider.zones)
	require.Empty(t, provider.recordIDs)
}

func TestDNSProvider_CleanUp_withoutPresent(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.no",
			servermock.ResponseFromInternal("get_domain.json")).
		Route("GET /domains/example.no/dns",
			servermock.ResponseFromInternal("list_records.json")).
		Route("DELETE /domains/example.no/dns/481523",
			servermock.ResponseFromInternal("delete_record.json")).
		Build(t)

	// keyAuth "456d==" hashes to the second record of the fixture.
	err := provider.CleanUp(t.Context(), "example.no", "abc", "456d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_recordNotFound(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.no",
			servermock.ResponseFromInternal("get_domain.json")).
		Route("GET /domains/example.no/dns",
			servermock.ResponseFromInternal("list_records.json")).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.no", "abc", "unknown")
	require.EqualError(t, err, "feno: record not found for '_acme-challenge.example.no.' 'sjpqhDnA3eVRWJPnyQweMjO4YW5jRHDyDcSSi882Cbw'")
}
