package gcore

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/internal/gcore/internal"
	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiToken string
		expected string
	}{
		{
			desc:     "success",
			apiToken: "A",
		},
		{
			desc:     "missing credentials",
			expected: "incomplete credentials provided",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.APIToken = test.apiToken

			p, err := NewDNSProviderConfig(config, "")

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

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := &Config{
				APIToken:   "secret",
				TTL:        120,
				HTTPClient: server.Client(),
			}

			p, err := NewDNSProviderConfig(config, "")
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("APIKey secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v2/zones/_acme-challenge.example.com",
			servermock.Noop().
				WithStatusCode(http.StatusNotFound),
		).
		Route("GET /v2/zones/example.com",
			servermock.ResponseFromInternal("get_zones.json"),
		).
		Route("GET /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.JSONEncode(internal.APIError{Message: "not found"}).
				WithStatusCode(http.StatusBadRequest),
		).
		Route("POST /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.ResponseFromInternal("create_rrset.json"),
			servermock.CheckRequestJSONBodyFromInternal("create_rrset-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v2/zones/_acme-challenge.example.com",
			servermock.Noop().
				WithStatusCode(http.StatusNotFound),
		).
		Route("GET /v2/zones/example.com",
			servermock.ResponseFromInternal("get_zones.json"),
		).
		Route("DELETE /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.Noop(),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
