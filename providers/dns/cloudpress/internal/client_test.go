package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "secret", "")
			if err != nil {
				return nil, err
			}

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
		Route("POST /api/dns_zones/z123/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:    RecordTypeTXT,
		Name:    "_acme-challenge.example.com",
		Value:   "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
		Comment: "Created by go-acme/lego",
	}

	result, err := client.CreateRecord(t.Context(), "z123", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "r123",
		Type:    RecordTypeTXT,
		Name:    "_acme-challenge.example.com",
		Value:   "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
		Comment: "Created by go-acme/lego",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/dns_zones/z123/records/r123",
			servermock.Noop().
				WithStatusCode(http.StatusAccepted),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "z123", "r123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/dns_zones/z123/records/r123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "z123", "r123")
	require.EqualError(t, err, "forbidden: An error message.")
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/dns_zones",
			servermock.ResponseFromFixture("list_dns_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("page", "1").
				With("per_page", "100"),
		).
		Build(t)

	pager := &Pager{
		Page:    1,
		PerPage: 100,
	}

	zones, err := client.ListZones(t.Context(), pager)
	require.NoError(t, err)

	expected := []Zone{
		{ID: "z123", Name: "example.com"},
		{ID: "z124", Name: "foo.example.com"},
		{ID: "z456", Name: "example.org"},
		{ID: "z457", Name: "example.net"},
	}

	assert.Equal(t, expected, zones)
}
