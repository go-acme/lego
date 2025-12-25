package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /zones/example.com/records/_acme-challenge.example.com/TXT",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
		).
		Build(t)

	request := RecordRequest{
		Zone:    "example.com",
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	records, err := client.AddRecord(t.Context(), request)
	require.NoError(t, err)

	expected := []Record{
		{Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY", TTL: 3600},
	}

	assert.Equal(t, expected, records)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/example.com/records/_acme-challenge.example.com/TXT",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckQueryParameter().Strict().
				With("content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"),
		).
		Build(t)

	request := RecordRequest{
		Zone:    "example.com",
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.DeleteRecord(t.Context(), request)
	require.NoError(t, err)
}
