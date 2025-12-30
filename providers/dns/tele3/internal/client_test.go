package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("keyA", "secretA")
			if err != nil {
				return nil, err
			}

			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.RawStringResponse("success"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
		).
		Build(t)

	err := client.AddRecord(t.Context(), "_acme-challenge.example.com", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.RawStringResponse("success"),
			servermock.CheckRequestJSONBodyFromFixture("remove_record-request.json"),
		).
		Build(t)

	err := client.RemoveRecord(t.Context(), "_acme-challenge.example.com", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)
}
