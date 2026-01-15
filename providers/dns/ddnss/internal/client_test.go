package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(&Authentication{Key: "secret"})
			if err != nil {
				return nil, err
			}

			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.ResponseFromFixture("success.html"),
			servermock.CheckQueryParameter().Strict().
				With("host", "_acme-challenge.example.com").
				With("key", "secret").
				With("txt", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("txtm", "1"),
		).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "_acme-challenge.example.com", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.ResponseFromFixture("success.html"),
			servermock.CheckQueryParameter().Strict().
				With("host", "_acme-challenge.example.com").
				With("key", "secret").
				With("txtm", "2"),
		).
		Build(t)

	err := client.RemoveTXTRecord(t.Context(), "_acme-challenge.example.com")
	require.NoError(t, err)
}
