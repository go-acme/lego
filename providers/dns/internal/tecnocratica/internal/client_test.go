package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock2.CheckHeader().WithJSONHeaders().
			With("X-TCpanel-Token", "secret"))
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones",
			servermock2.ResponseFromFixture("get_zones.json")).
		Build(t)

	zones, err := client.GetZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:        6,
			Name:      "example.com",
			HumanName: "example.com",
		},
		{
			ID:        7,
			Name:      "example.org",
			HumanName: "example.org",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones",
			servermock2.RawStringResponse(`{"error": "unauthorized"}`).
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	zones, err := client.GetZones(t.Context())
	require.Error(t, err)

	assert.Nil(t, zones)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones/6/records",
			servermock2.ResponseFromFixture("get_records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), 6, "")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      98,
			Name:    "",
			Type:    "SOA",
			Content: "ns1.example.org dns.example.org 2015092102 7200 7200 1209600 1800",
			TTL:     7200,
		},
		{
			ID:      99,
			Name:    "",
			Type:    "NS",
			Content: "ns1.example.org",
			TTL:     7200,
		},
		{
			ID:      100,
			Name:    "_acme-challenge",
			Type:    "TXT",
			Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			TTL:     120,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/zones/6/records",
			servermock2.ResponseFromFixture("create_record.json").
				WithStatusCode(http.StatusCreated),
			servermock2.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	result, err := client.CreateRecord(t.Context(), 6, record)
	require.NoError(t, err)

	expected := &Record{
		ID:      101,
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/zones/6/records",
			servermock2.RawStringResponse(`{"error": "bad request"}`).
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	record := Record{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "test-value",
		TTL:     120,
	}

	result, err := client.CreateRecord(t.Context(), 6, record)
	require.Error(t, err)

	assert.Nil(t, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/zones/6/records/101",
			servermock2.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteRecord(t.Context(), 6, 101)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/zones/6/records/999",
			servermock2.RawStringResponse(`{"error": "not found"}`).
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	err := client.DeleteRecord(t.Context(), 6, 999)
	require.Error(t, err)
}
