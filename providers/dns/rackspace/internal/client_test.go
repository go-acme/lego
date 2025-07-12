package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
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
		servermock.CheckHeader().WithJSONHeaders().
			With(AuthToken, "secret"))
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/1234/records",
			servermock.ResponseFromFixture("add-records.json"),
			servermock.CheckRequestJSONBody(`{"records":[{"name":"exmaple.com","type":"TXT","data":"value1","ttl":120,"id":"abc"}]}`)).
		Build(t)

	record := Record{
		Name: "exmaple.com",
		Type: "TXT",
		Data: "value1",
		TTL:  120,
		ID:   "abc",
	}

	err := client.AddRecord(t.Context(), "1234", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/1234/records", nil).
		Build(t)

	err := client.DeleteRecord(t.Context(), "1234", "2725233")
	require.NoError(t, err)
}

func TestClient_searchRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/1234/records", servermock.ResponseFromFixture("search-records.json")).
		Build(t)

	records, err := client.searchRecords(t.Context(), "1234", "2725233", "A")
	require.NoError(t, err)

	expected := &Records{
		TotalEntries: 6,
		Records: []Record{
			{Name: "ftp.example.com", Type: "A", Data: "192.0.2.8", TTL: 5771, ID: "A-6817754"},
			{Name: "example.com", Type: "A", Data: "192.0.2.17", TTL: 86400, ID: "A-6822994"},
			{Name: "example.com", Type: "NS", Data: "ns.rackspace.com", TTL: 3600, ID: "NS-6251982"},
			{Name: "example.com", Type: "NS", Data: "ns2.rackspace.com", TTL: 3600, ID: "NS-6251983"},
			{Name: "example.com", Type: "MX", Data: "mail.example.com", TTL: 3600, ID: "MX-3151218"},
			{Name: "www.example.com", Type: "CNAME", Data: "example.com", TTL: 5400, ID: "CNAME-9778009"},
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_listDomainsByName(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains", servermock.ResponseFromFixture("list-domains-by-name.json")).
		Build(t)

	domains, err := client.listDomainsByName(t.Context(), "1234")
	require.NoError(t, err)

	expected := &ZoneSearchResponse{
		TotalEntries: 114,
		HostedZones:  []HostedZone{{ID: "2725257", Name: "sub1.example.com"}},
	}

	assert.Equal(t, expected, domains)
}
