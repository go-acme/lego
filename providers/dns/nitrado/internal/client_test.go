package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
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

func TestClient_InsertRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/example.com/records",
			servermock.ResponseFromFixture("insert_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("insert_record-request.json"),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     "120",
	}

	err := client.InsertRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_InsertRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     "120",
	}

	err := client.InsertRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "error: Access token not valid. access_token_not_valid: {}")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domain/example.com/records",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("delete_record-request.json"),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}
