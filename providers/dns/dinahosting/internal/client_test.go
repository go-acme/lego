package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckQueryParameter().
			With("AUTH_USER", "user").
			With("AUTH_PWD", "secret").
			With("responseType", "Json"),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.Noop(),
			servermock.CheckQueryParameter().
				With("domain", "example.com").
				With("hostname", "_acme-challenge").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("command", "Domain_Zone_AddTypeTXT"),
		).
		Build(t)

	record := TXTRecord{
		Domain:   "example.com",
		Hostname: "_acme-challenge",
		Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.AddTXTRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	record := TXTRecord{
		Domain:   "example.com",
		Hostname: "_acme-challenge",
		Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.AddTXTRecord(t.Context(), record)
	require.EqualError(t, err, "2200: Domain_Zone_AddTypeTXT (dh6958820711fcf9.04537876): 2200: Authentication error.")
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.DumpRequest(),
			servermock.CheckQueryParameter().
				With("domain", "example.com").
				With("hostname", "_acme-challenge").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("command", "Domain_Zone_DeleteTypeTXT"),
		).
		Build(t)

	record := TXTRecord{
		Domain:   "example.com",
		Hostname: "_acme-challenge",
		Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.DeleteTXTRecord(t.Context(), record)
	require.NoError(t, err)
}
