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
			With(APIKeyHeader, "secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/domain/example.com/entry_set",
			servermock.ResponseFromFixture("add_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
		).
		Build(t)

	record := Record{
		Host:  "_acme-challenge",
		Type:  "TXT",
		TTL:   120,
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:    1,
		Host:  "_acme-challenge",
		Type:  "TXT",
		TTL:   120,
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/domain/example.com/entry_set/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.NoError(t, err)
}
