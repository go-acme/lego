package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder(hc *http.Client) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(hc)
			client.APIEndpoint, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer xxx"),
	)
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("POST /zones/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("post_zoneszonerecords.json"),
			servermock.CheckRequestJSONBodyFromFixture("post_zoneszonerecords-request.json"),
		).
		Build(t)

	err := client.CreateTXTRecord(t.Context(), "example.com", "_acme-challenge", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY", 120)
	require.NoError(t, err)
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("DELETE /zones/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("delete_zoneszonerecords.json"),
			servermock.CheckQueryParameter().Strict().
				With("data", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		).
		Build(t)

	err := client.RemoveTXTRecord(t.Context(), "example.com", "_acme-challenge", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)
}
