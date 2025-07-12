package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder(apiToken string) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("me", apiToken)
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("me", apiToken))
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder("secretA").
		Route("GET /v1/zones", servermock.ResponseFromFixture("list_zones.json")).
		Build(t)

	zones, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []DNSZone{
		{
			ID:             1,
			Name:           "example.com",
			Synced:         false,
			QueriesCount:   0,
			RecordsCount:   3,
			AliasesCount:   0,
			RedirectsCount: 0,
			ForwardsCount:  0,
			TemplateID:     0,
		},
		{
			ID:             2,
			Name:           "example.net",
			Synced:         false,
			QueriesCount:   0,
			RecordsCount:   3,
			AliasesCount:   0,
			RedirectsCount: 0,
			ForwardsCount:  0,
			TemplateID:     0,
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder("secretB").
		Route("POST /v1/zones/1/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBody(`{"name":"example.com.","type":"MX","content":"10 mail.example.com.","ttl":300}`)).
		Build(t)

	zone := DNSZone{ID: 1}

	record := DNSRecord{
		Name:    "example.com.",
		Type:    "MX",
		Content: "10 mail.example.com.",
		TTL:     300,
	}

	newRecord, err := client.CreateRecord(t.Context(), zone, record)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:      100,
		Name:    "example.com.",
		Type:    "MX",
		Content: "10 mail.example.com.",
		TTL:     300,
		ZoneID:  1,
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder("secretC").
		Route("DELETE /v1/zones/1/records/2",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckRequestJSONBody(`{"id":2,"name":"example.com.","type":"MX","content":"10 mail.example.com.","ttl":300,"zone_id":1}`)).
		Build(t)

	record := &DNSRecord{
		ID:      2,
		Name:    "example.com.",
		Type:    "MX",
		Content: "10 mail.example.com.",
		TTL:     300,
		ZoneID:  1,
	}

	err := client.DeleteRecord(t.Context(), record)
	require.NoError(t, err)
}
