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

func setupClient(token string) func(server *httptest.Server) (*Client, error) {
	return func(server *httptest.Server) (*Client, error) {
		client := NewClient(OAuthStaticAccessToken(server.Client(), token))
		client.baseURL, _ = url.Parse(server.URL)

		return client, nil
	}
}

func TestClient_GetRecords(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient("tokenA"),
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer tokenA"),
	).
		Route("GET /dns_zones/zoneID/dns_records",
			servermock.ResponseFromFixture("get_records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "zoneID")
	require.NoError(t, err)

	expected := []DNSRecord{
		{ID: "u6b433c15a27a2d79c6616d6", Hostname: "example.org", TTL: 3600, Type: "A", Value: "10.10.10.10"},
		{ID: "u6b4764216f272872ac0ff71", Hostname: "test.example.org", TTL: 300, Type: "TXT", Value: "txtxtxtxtxtxt"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient("tokenB"),
		servermock.CheckHeader().
			WithAccept("application/json").
			WithContentType("application/json; charset=utf-8").
			WithAuthorization("Bearer tokenB"),
	).
		Route("POST /dns_zones/zoneID/dns_records",
			servermock.ResponseFromFixture("create_record.json").
				WithStatusCode(http.StatusCreated)).
		Build(t)

	record := DNSRecord{
		Hostname: "_acme-challenge.example.com",
		TTL:      300,
		Type:     "TXT",
		Value:    "txtxtxtxtxtxt",
	}

	result, err := client.CreateRecord(t.Context(), "zoneID", record)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:       "u6b4764216f272872ac0ff71",
		Hostname: "test.example.org",
		TTL:      300,
		Type:     "TXT",
		Value:    "txtxtxtxtxtxt",
	}

	assert.Equal(t, expected, result)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient("tokenC"),
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer tokenC"),
	).
		Route("DELETE /dns_zones/zoneID/dns_records/recordID",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.RemoveRecord(t.Context(), "zoneID", "recordID")
	require.NoError(t, err)
}
