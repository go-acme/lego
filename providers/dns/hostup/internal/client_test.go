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
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns-zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("limit", "1000").
				With("name", "example.com"),
		).
		Build(t)

	zones, err := client.ListZones(t.Context(), "example.com", 1000, "")
	require.NoError(t, err)

	expected := []Zone{{
		ID:         "zone_01hxa3b4c5d6e7f8g9h0j1k2m3",
		Name:       "example.com",
		ZoneStatus: "active",
		IsDNSOnly:  false,
		Kind:       "linked",
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns-zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	_, err := client.ListZones(t.Context(), "example.com", 1000, "")
	require.EqualError(t, err,
		"400: Invalid request The request body failed validation. invalid_request"+
			" (/api/v2/resource - req_01hxa3b4c5d6e7f8g9h0j1k2m3)"+
			" [invalid_request: `domainName` is required. /items/0/domainName]",
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns-zones/zone_01hxa3b4c5d6e7f8g9h0j1k2m3/records",
			servermock.ResponseFromFixture("record_create.json"),
			servermock.CheckRequestJSONBodyFromFixture("record_create-request.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   120,
	}

	result, err := client.CreateRecord(t.Context(), "zone_01hxa3b4c5d6e7f8g9h0j1k2m3", record)
	require.NoError(t, err)

	expected := &Record{
		ID:     "drr_01hxa3b4c5d6e7f8g9h0j1k2m3",
		Type:   "TXT",
		Name:   "_acme-challenge.example.com",
		Value:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:    120,
		Status: "pending",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns-zones/zone_01hxa3b4c5d6e7f8g9h0j1k2m3/records/drr_01hxa3b4c5d6e7f8g9h0j1k2m3",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "zone_01hxa3b4c5d6e7f8g9h0j1k2m3", "drr_01hxa3b4c5d6e7f8g9h0j1k2m3")
	require.NoError(t, err)
}
