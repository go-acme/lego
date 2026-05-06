package active24

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		secret   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "user",
			secret: "secret",
		},
		{
			desc:     "missing API key",
			apiKey:   "",
			secret:   "secret",
			expected: "credentials missing",
		},
		{
			desc:     "missing secret",
			apiKey:   "user",
			secret:   "",
			expected: "credentials missing",
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
			config.Secret = test.secret

			p, err := NewDNSProviderConfig(config, "example.com")

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
				APIKey:     "user",
				Secret:     "secret",
				TTL:        120,
				HTTPClient: server.Client(),
			}

			p, err := NewDNSProviderConfig(config, "example.com")
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithRegexp("Authorization", `Basic .+`).
			WithRegexp("Date", `\d+-\d+-\d+T\d{2}:\d{2}:\d{2}.*`).
			With("Accept-Language", "en_us"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v1/user/self/service",
			servermock.ResponseFromInternal("services.json"),
		).
		Route("POST /v2/service/3333/dns/record",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v1/user/self/service",
			servermock.ResponseFromInternal("services.json"),
		).
		Route("GET /v2/service/3333/dns/record",
			servermock.ResponseFromInternal("records.json"),
		).
		Route("DELETE /v2/service/3333/dns/record/14",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
