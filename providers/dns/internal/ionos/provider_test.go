package ionos

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/internal/ionos/internal"
	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		tll      int
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
			tll:    MinTTL,
		},
		{
			desc:     "missing credentials",
			tll:      MinTTL,
			expected: "credentials missing",
		},
		{
			desc:     "invalid TTL",
			apiKey:   "123",
			tll:      30,
			expected: "invalid TTL, TTL (30) must be greater than 300",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.APIKey = test.apiKey
			config.TTL = test.tll

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
				TTL:        300,
				HTTPClient: server.Client(),
			}

			p, err := NewDNSProviderConfig(config, server.URL)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With(internal.APIKeyHeader, "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v1/zones",
			servermock.ResponseFromInternal("list_zones.json"),
		).
		Route("GET /v1/zones/f6821d2f-2ff3-4762-99bd-711733624a77",
			servermock.ResponseFromInternal("get_records.json"),
			servermock.CheckQueryParameter().Strict().
				With("recordType", "TXT").
				With("suffix", "_acme-challenge.example.com"),
		).
		Route("PATCH /v1/zones/f6821d2f-2ff3-4762-99bd-711733624a77",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromInternal("replace_records-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v1/zones",
			servermock.ResponseFromInternal("list_zones.json"),
		).
		Route("GET /v1/zones/f6821d2f-2ff3-4762-99bd-711733624a77",
			servermock.ResponseFromInternal("get_records-remove.json"),
			servermock.CheckQueryParameter().Strict().
				With("recordType", "TXT").
				With("suffix", "_acme-challenge.example.com"),
		).
		Route("DELETE /v1/zones/f6821d2f-2ff3-4762-99bd-711733624a77/records/830a8feb-fc36-4641-980b-043d4402d34e",
			servermock.Noop(),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
