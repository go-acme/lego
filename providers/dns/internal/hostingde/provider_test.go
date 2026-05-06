package hostingde

import (
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
		zoneName string
		expected string
	}{
		{
			desc:     "success",
			apiKey:   "123",
			zoneName: "example.org",
		},
		{
			desc:     "missing credentials",
			expected: "API key missing",
		},
		{
			desc:     "missing api key",
			zoneName: "456",
			expected: "API key missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.APIKey = test.apiKey
			config.ZoneName = test.zoneName

			p, err := NewDNSProviderConfig(config, "")

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
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

			p, err := NewDNSProviderConfig(config, server.URL)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zoneConfigsFind",
			servermock.ResponseFromInternal("zoneConfigsFind.json"),
			servermock.CheckRequestJSONBodyFromInternal("zoneConfigsFind-request.json"),
		).
		Route("POST /zoneUpdate",
			servermock.ResponseFromInternal("zoneUpdate.json"),
			servermock.CheckRequestJSONBodyFromInternal("zoneUpdate-request_add.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /zoneConfigsFind",
			servermock.ResponseFromInternal("zoneConfigsFind.json"),
			servermock.CheckRequestJSONBodyFromInternal("zoneConfigsFind-request.json"),
		).
		Route("POST /zoneUpdate",
			servermock.ResponseFromInternal("zoneUpdate.json"),
			servermock.CheckRequestJSONBodyFromInternal("zoneUpdate-request_remove.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
