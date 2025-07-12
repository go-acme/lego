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
			client := NewClient("secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains", servermock.ResponseFromFixture("domain-list-domains.json")).
		Build(t)

	domains, err := client.ListDomains(t.Context())
	require.NoError(t, err)

	require.Len(t, domains, 1)

	expected := []Domain{{
		Domain:    "string",
		DomainID:  "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		ProjectID: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
	}}

	assert.Equal(t, expected, domains)
}

func TestClient_ListDomains_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains",
			servermock.ResponseFromFixture("error-client.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.ListDomains(t.Context())
	require.EqualError(t, err, "[status code 400] ValidationError: Validation failed [format: should be string (.address.street, email)]")
}

func TestClient_ListDNSZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /projects/my-project-id/dns-zones", servermock.ResponseFromFixture("dns-list-dns-zones.json")).
		Build(t)

	zones, err := client.ListDNSZones(t.Context(), "my-project-id")
	require.NoError(t, err)

	require.Len(t, zones, 1)

	expected := []DNSZone{{
		ID:     "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		Domain: "string",
		RecordSet: &RecordSet{
			TXT: &TXTRecord{},
		},
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_GetDNSZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns-zones/my-zone-id", servermock.ResponseFromFixture("dns-get-dns-zone.json")).
		Build(t)

	zone, err := client.GetDNSZone(t.Context(), "my-zone-id")
	require.NoError(t, err)

	expected := &DNSZone{
		ID:     "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		Domain: "string",
		RecordSet: &RecordSet{
			TXT: &TXTRecord{},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_CreateDNSZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns-zones",
			servermock.ResponseFromFixture("dns-create-dns-zone.json"),
			servermock.CheckRequestJSONBody(`{"name":"test","parentZoneId":"my-parent-zone-id"}`)).
		Build(t)

	request := CreateDNSZoneRequest{
		Name:         "test",
		ParentZoneID: "my-parent-zone-id",
	}

	zone, err := client.CreateDNSZone(t.Context(), request)
	require.NoError(t, err)

	expected := &DNSZone{
		ID: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_UpdateTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /dns-zones/my-zone-id/record-sets/txt",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBody(`{"settings":{"ttl":{"auto":true}},"entries":["txt"]}`)).
		Build(t)

	record := TXTRecord{
		Settings: Settings{
			TTL: TTL{Auto: true},
		},
		Entries: []string{"txt"},
	}

	err := client.UpdateTXTRecord(t.Context(), "my-zone-id", record)
	require.NoError(t, err)
}

func TestClient_DeleteDNSZone(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns-zones/my-zone-id",
			servermock.Noop()).
		Build(t)

	err := client.DeleteDNSZone(t.Context(), "my-zone-id")
	require.NoError(t, err)
}

func TestClient_DeleteDNSZone_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns-zones/my-zone-id",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	err := client.DeleteDNSZone(t.Context(), "my-zone-id")
	assert.EqualError(t, err, "[status code 500] InternalServerError: Something went wrong")
}
