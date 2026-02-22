package leaseweb

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/platform/tester"
	"github.com/go-acme/lego/v5/platform/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/leaseweb/internal"
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
			expected: "leaseweb: some credentials information are missing: LEASEWEB_API_KEY",
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
			desc:     "missing credentials",
			expected: "leaseweb: credentials missing",
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
			With(internal.AuthHeader, "secret"),
	)
}

func TestDNSProvider_Present_create(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromInternal("error_404.json").
				WithStatusCode(http.StatusNotFound),
		).
		Route("POST /domains/example.com/resourceRecordSets",
			servermock.ResponseFromInternal("createResourceRecordSet.json"),
			servermock.CheckRequestJSONBodyFromInternal("createResourceRecordSet-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_update(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromInternal("getResourceRecordSet.json"),
		).
		Route("PUT /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromInternal("updateResourceRecordSet.json"),
			servermock.CheckRequestJSONBodyFromInternal("updateResourceRecordSet-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_delete(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromInternal("getResourceRecordSet2.json"),
		).
		Route("DELETE /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "1234d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_update(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromInternal("getResourceRecordSet.json"),
		).
		Route("PUT /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromInternal("updateResourceRecordSet.json"),
			servermock.CheckRequestJSONBodyFromInternal("updateResourceRecordSet-request2.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "1234d==")
	require.NoError(t, err)
}
