package internal

import (
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
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /v1/zones/zone_abc123/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.example.com",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	result, err := client.CreateRecord(t.Context(), "zone_abc123", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "rec_abc123",
		Name:    "_acme-challenge.example.com.",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/zones/zone_abc123/records/rec_abc123",
			servermock.Noop(),
		).
		Build(t)

	err := client.DeleteRecordByID(t.Context(), "zone_abc123", "rec_abc123")
	require.NoError(t, err)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones",
			servermock.ResponseFromFixture("list_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("limit", "100").
				With("offset", "0"),
		).
		Build(t)

	zones, err := client.ListZones(t.Context(), &Pager{Limit: 100})
	require.NoError(t, err)

	expected := &PaginatedData[Zone]{
		Items: []Zone{
			{ID: "zone_ghi789", Name: "foo.example.com", Type: "primary", Status: "active", RecordCount: 6},
			{ID: "zone_abc123", Name: "example.com", Type: "primary", Status: "active", RecordCount: 15, DNSSecEnabled: true},
			{ID: "zone_def456", Name: "example.org", Type: "primary", Status: "active", RecordCount: 8},
		},
		Total:  2,
		Limit:  20,
		Offset: 0,
	}

	assert.Equal(t, expected, zones)
}
