package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
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

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/v2/Domains/example.com/Records",
			servermock.Noop().
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("records_create-request.json"),
		).
		Build(t)

	record := Record{
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/v2/Domains/example.com/Records/1234",
			servermock.Noop().
				WithStatusCode(http.StatusOK),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 1234)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/v2/Domains/example.com/Records/1234",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 1234)
	require.EqualError(t, err, "type: string, title: string, detail: string, instance: string")
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/v2/Domains/example.com/Records",
			servermock.ResponseFromFixture("records_get.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT"),
		).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com", "_acme-challenge", "TXT")
	require.NoError(t, err)

	expected := []Record{{
		ID:   1234,
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}}

	assert.Equal(t, expected, records)
}
