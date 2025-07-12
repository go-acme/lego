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
			client := NewClient("user", "secret")
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_ListZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /dnszones/",
			servermock.ResponseFromFixture("ListZone.json")).
		Build(t)

	ctx := t.Context()

	zones, err := client.ListZone(ctx)
	require.NoError(t, err)

	expected := []Zone{
		{Name: "example.com", Type: "master"},
		{Name: "example.net", Type: "slave"},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZoneDetails(t *testing.T) {
	client := mockBuilder().
		Route("GET /dnszones/example.com",
			servermock.ResponseFromFixture("GetZoneDetails.json")).
		Build(t)

	zone, err := client.GetZoneDetails(t.Context(), "example.com")
	require.NoError(t, err)

	expected := &ZoneDetails{
		Active: true,
		DNSSec: true,
		Name:   "example.com",
		Type:   "master",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_ListRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dnszones/example.com/records/",
			servermock.ResponseFromFixture("ListRecords.json")).
		Build(t)

	records, err := client.ListRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			Name: "@",
			TTL:  86400,
			Type: "NS",
			Data: "ns2.core-networks.eu.",
		},
		{
			Name: "@",
			TTL:  86400,
			Type: "NS",
			Data: "ns3.core-networks.com.",
		},
		{
			Name: "@",
			TTL:  86400,
			Type: "NS",
			Data: "ns1.core-networks.de.",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dnszones/example.com/records/",
			servermock.Noop().WithStatusCode(http.StatusNoContent)).
		Build(t)

	record := Record{Name: "www", TTL: 3600, Type: "A", Data: "127.0.0.1"}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /dnszones/example.com/records/delete",
			servermock.Noop().WithStatusCode(http.StatusNoContent)).
		Build(t)

	record := Record{Name: "www", Type: "A", Data: "127.0.0.1"}

	err := client.DeleteRecords(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_CommitRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /dnszones/example.com/records/commit",
			servermock.Noop().WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.CommitRecords(t.Context(), "example.com")
	require.NoError(t, err)
}
