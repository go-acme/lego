package digitalocean

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAuthToken)

func mockProvider() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.AuthToken = "asdf1234"
			config.BaseURL = server.URL
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("Authorization", "Bearer asdf1234"),
	)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAuthToken: "123",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAuthToken: "",
			},
			expected: "digitalocean: some credentials information are missing: DO_AUTH_TOKEN",
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
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		authToken string
		expected  string
	}{
		{
			desc:      "success",
			authToken: "123",
		},
		{
			desc:     "missing credentials",
			expected: "digitalocean: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthToken = test.authToken

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockProvider().
		Route("POST /v2/domains/example.com/records",
			servermock.RawStringResponse(`{
			"domain_record": {
				"id": 1234567,
				"type": "TXT",
				"name": "_acme-challenge",
				"data": "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
				"priority": null,
				"port": null,
				"weight": null
			}
		}`).
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBody(`{"type":"TXT","name":"_acme-challenge.example.com.","data":"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY","ttl":30}`)).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]int{"abc": 1234567}, provider.recordIDs)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockProvider().
		Route("DELETE /v2/domains/example.com/records/1234567",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	provider.recordIDsMu.Lock()
	provider.recordIDs["abc"] = 1234567
	provider.recordIDsMu.Unlock()

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.recordIDs)
}
