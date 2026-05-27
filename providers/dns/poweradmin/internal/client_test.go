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
			client, err := NewClient(server.URL, "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With(AuthenticationHeader, "secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/v2/zones/1/records",
			servermock.ResponseFromFixture("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	r := Record{
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	record, err := client.CreateRecord(t.Context(), 1, r)
	require.NoError(t, err)

	expected := &Record{
		ID:      456,
		ZoneID:  1,
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/v2/zones/1/records/456",
			servermock.ResponseFromFixture("delete_record.json").
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 1, 456)
	require.NoError(t, err)
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v2/zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("page", "1").
				With("per_page", "25"),
		).
		Build(t)

	pager := &Pager{
		Page:    1,
		PerPage: 25,
	}

	zones, pagination, err := client.GetZones(t.Context(), pager)
	require.NoError(t, err)

	expected := []Zone{
		{ID: 1, Name: "example.com", Type: "MASTER", CreatedAt: "2025-01-01 12:00:00"},
		{ID: 2, Name: "10.in-addr.arpa", Type: "NATIVE"},
		{ID: 3, Name: "example.org", Type: "NATIVE"},
		{ID: 4, Name: "foo.example.com", Type: "MASTER", CreatedAt: "2025-01-01 12:00:00"},
		{ID: 5, Name: "bar.foo.example.com", Type: "MASTER", CreatedAt: "2025-01-01 12:00:00"},
	}

	assert.Equal(t, expected, zones)

	expectedPagination := &Pagination{CurrentPage: 1, PerPage: 100, Total: 4, LastPage: 1}

	assert.Equal(t, expectedPagination, pagination)
}
