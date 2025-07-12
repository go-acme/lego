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

const apiKey = "key"

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(OAuthStaticAccessToken(server.Client(), apiKey))
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer "+apiKey))
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v1/zones/example.com/dns-records", servermock.ResponseFromFixture("RecordsResponse.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:   "string",
			Type: "string",
			Name: "string",
			Contents: []Content{
				{
					Text: "string",
				},
			},
			TTL: 3600,
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_GetRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v1/zones/example.com/dns-records/123", servermock.ResponseFromFixture("RecordResponse.json")).
		Build(t)

	record, err := client.GetRecord(t.Context(), "example.com", "123")
	require.NoError(t, err)

	expected := &Record{
		ID:   "string",
		Type: "string",
		Name: "string",
		Contents: []Content{
			{
				Text: "string",
			},
		},
		TTL: 3600,
	}
	assert.Equal(t, expected, record)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/v1/zones/example.com/dns-records",
			servermock.ResponseFromFixture("RecordResponse.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBody(`{"name":"string","type":"string","ttl":3600,"contents":[{"text":"string"}]}`)).
		Build(t)

	data := Record{
		Type: "string",
		Name: "string",
		Contents: []Content{
			{
				Text: "string",
			},
		},
		TTL: 3600,
	}

	record, err := client.CreateRecord(t.Context(), "example.com", data)
	require.NoError(t, err)

	expected := &Record{
		ID:   "string",
		Type: "string",
		Name: "string",
		Contents: []Content{
			{
				Text: "string",
			},
		},
		TTL: 3600,
	}

	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/v1/zones/example.com/dns-records/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_NotFound_Response(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/v1/zones/example.com/dns-records/123",
			servermock.Noop().
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/v1/zones/example.com/dns-records/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "123")
	require.EqualError(t, err, "[status code: 401] Unauthorized: Invalid token missing header")
}
