package curanet

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

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
			expected: "credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.APIKey = test.apiKey

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
				APIKey:     "secret",
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
			WithAuthorization("Bearer secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /dns/v2/Domains/example.com/Records",
			servermock.Noop().
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("records_create-request.json"),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/v2/Domains/example.com/Records",
			servermock.ResponseFromInternal("records_get.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT"),
		).
		Route("DELETE /dns/v2/Domains/example.com/Records/1234",
			servermock.Noop().
				WithStatusCode(http.StatusOK),
		).
		Build(t)

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
