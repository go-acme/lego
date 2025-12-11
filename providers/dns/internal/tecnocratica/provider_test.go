package tecnocratica

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		token    string
		expected string
	}{
		{
			desc:  "success",
			token: "secret",
		},
		{
			desc:     "missing token",
			expected: "missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.Token = test.token

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
				Token:              "secret",
				PropagationTimeout: 10 * time.Second,
				PollingInterval:    1 * time.Second,
				TTL:                120,
				HTTPClient:         server.Client(),
			}

			p, err := NewDNSProviderConfig(config, server.URL)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With("X-TCpanel-Token", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /dns/zones",
			servermock.ResponseFromInternal("get_zones.json")).
		Route("POST /dns/zones/6/records",
			servermock.ResponseFromInternal("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromInternal("create_record-request.json")).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /dns/zones/456/records/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	token := "abc"

	provider.recordIDs[token] = 123
	provider.zoneIDs[token] = 456

	err := provider.CleanUp("example.com", token, "123d==")
	require.NoError(t, err)
}
