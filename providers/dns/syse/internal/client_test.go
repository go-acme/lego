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
			client, err := NewClient(map[string]string{
				"example.com": "secret",
			})
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/example.com",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := Record{
		Type:    "TXT",
		Prefix:  "_acme-challenge",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Active:  true,
		TTL:     120,
	}

	result, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "1234",
		Type:    "TXT",
		Prefix:  "_acme-challenge",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Active:  true,
		TTL:     120,
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/example.com",
			servermock.RawStringResponse(http.StatusText(http.StatusUnauthorized)).
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Type:    "TXT",
		Prefix:  "_acme-challenge",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Active:  true,
		TTL:     120,
	}

	_, err := client.CreateRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "unexpected status code: [status code: 401] body: Unauthorized")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/example.com/1234",
			servermock.Noop()).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "1234")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/example.com/1234",
			servermock.RawStringResponse(http.StatusText(http.StatusUnauthorized)).
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "1234")
	require.EqualError(t, err, "unexpected status code: [status code: 401] body: Unauthorized")
}
