package curanet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc         string
		apiKey       string
		clientID     string
		clientSecret string
		expected     string
	}{
		{
			desc:   "success (deprecated)",
			apiKey: "secret",
		},
		{
			desc:         "success",
			clientID:     "xxx",
			clientSecret: "secret",
		},
		{
			desc:     "success",
			clientID: "xxx",
			expected: "credentials missing",
		},
		{
			desc:         "success",
			clientSecret: "secret",
			expected:     "credentials missing",
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
			config.ClientID = test.clientID
			config.ClientSecret = test.clientSecret

			p, err := NewDNSProviderConfig(config, URLConfiguration{})

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

func mockBuilder(tokenURL string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := &Config{
				TTL:          120,
				ClientID:     "123",
				ClientSecret: "xxx",
			}

			urlCfg := URLConfiguration{
				BaseURL:  server.URL,
				TokenURL: tokenURL,
			}

			p, err := NewDNSProviderConfig(config, urlCfg)
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

func oauthServer(t *testing.T) string {
	t.Helper()

	return servermock.NewBuilder[string](
		func(server *httptest.Server) (string, error) {
			return server.URL, nil
		}).
		Route("POST /",
			servermock.ResponseFromInternal("oauth.json"),
			servermock.CheckForm().Strict().UsePostForm().
				With("client_id", "123").
				With("client_secret", "xxx").
				With("grant_type", "client_credentials").
				With("scope", "dns"),
		).
		Build(t)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("POST /dns/v2/Domains/example.com/Records",
			servermock.Noop().
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("records_create-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("GET /dns/v2/Domains/example.com/Records",
			servermock.ResponseFromInternal("records_get.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge.example.com").
				With("type", "TXT"),
		).
		Route("DELETE /dns/v2/Domains/example.com/Records/2373303",
			servermock.Noop().
				WithStatusCode(http.StatusOK),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestCreateURLConfiguration(t *testing.T) {
	cfg := CreateURLConfiguration("foo")

	expected := URLConfiguration{
		BaseURL:  "https://api.foo.dk/",
		TokenURL: "https://apiauth.dk.team.blue/realms/foo/protocol/openid-connect/token",
	}

	assert.Equal(t, expected, cfg)
}
