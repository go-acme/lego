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
			client, err := NewClient("key", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With("X-Api-Key", "key").
			With("X-Api-Secret", "secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /dns/records/example.com", nil,
			servermock.CheckRequestJSONBody(`{"items":[{"type":"TXT","name":"@","ttl":60}]}`)).
		Build(t)

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /dns/records/example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnprocessableEntity)).
		Build(t)

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "^$, name: The domain name contains invalid characters")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/records/example.com", nil,
			servermock.CheckRequestJSONBody(`[{"type":"TXT","name":"@","ttl":60}]`)).
		Build(t)

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/records/example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnprocessableEntity)).
		Build(t)

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "^$, name: The domain name contains invalid characters")
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/records/example.com",
			servermock.ResponseFromFixture("get-records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{Type: "A", Name: "@", TTL: 3600},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/records/example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnprocessableEntity)).
		Build(t)

	_, err := client.GetRecords(t.Context(), "example.com")
	require.EqualError(t, err, "^$, name: The domain name contains invalid characters")
}
