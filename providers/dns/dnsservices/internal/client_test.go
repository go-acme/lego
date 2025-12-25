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
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("user", "secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/123/dns/456/records",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		TTL:     120,
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.AddRecord(t.Context(), "123", "456", record)
	require.NoError(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /service/123/dns/456/records/789",
			servermock.Noop(),
		).
		Build(t)

	err := client.RemoveRecord(t.Context(), "123", "456", "789")
	require.NoError(t, err)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns",
			servermock.ResponseFromFixture("zones.json"),
		).
		Build(t)

	zones, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{DomainID: "60", Name: "example.com", ServiceID: "10"},
		{DomainID: "61", Name: "example.org", ServiceID: "20"},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	_, err := client.ListZones(t.Context())
	require.EqualError(t, err, "wronglogin: unauthorized")
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /service/123/dns/456",
			servermock.ResponseFromFixture("records.json"),
		).
		Build(t)

	records, err := client.GetRecords(t.Context(), "123", "456")
	require.NoError(t, err)

	expected := []Record{
		{ID: "10", Name: "qwerty", TTL: 1800, Type: "A", Content: "127.0.0.1"},
		{ID: "11", Name: "qwerty", TTL: 1800, Type: "NS", Content: "ns1.example.com"},
		{ID: "66", Name: "_acme-challenge", TTL: 120, Type: "TXT", Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
	}

	assert.Equal(t, expected, records)
}
