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

func mockBuilder(httpClient *http.Client) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(httpClient)
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("POST /dns/v2/Domains/example.com/Records",
			servermock.Noop().
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("records_create-request.json"),
		).
		Build(t)

	record := Record{
		Name: "_acme-challenge.example.com",
		Type: "TXT",
		TTL:  120,
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("DELETE /dns/v2/Domains/example.com/Records/1234",
			servermock.Noop().
				WithStatusCode(http.StatusOK),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 1234)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("DELETE /dns/v2/Domains/example.com/Records/1234",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 1234)
	require.EqualError(t, err, "type: https://httpstatuses.com/400, title: Validation error(s), status: 400, detail: An error occurred during validation, see errors for details")
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("GET /dns/v2/Domains/example.com/Records",
			servermock.ResponseFromFixture("records_get.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT"),
		).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com", "_acme-challenge", "TXT")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:   2373303,
			Name: "_acme-challenge.example.com",
			Type: "TXT",
			TTL:  120,
			Data: `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`,
		},
		{
			ID:   2373304,
			Name: "_acme-challenge.example.com",
			Type: "TXT",
			TTL:  120,
			Data: `"abc"`,
		},
	}

	assert.Equal(t, expected, records)
}
