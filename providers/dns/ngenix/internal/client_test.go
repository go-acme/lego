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
			client, err := NewClient("user", "secret", "42")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("user/token", "secret"),
	)
}

func TestClient_ListDNSZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns-zone",
			servermock.ResponseFromFixture("list_dns_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("customerId", "42"),
		).
		Build(t)

	zones, err := client.ListDNSZones(t.Context())
	require.NoError(t, err)

	expected := []DNSZoneCollectionView{
		{ID: 1, Name: "example.com"},
		{ID: 2, Name: "example.org"},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetDNSZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns-zone/1",
			servermock.ResponseFromFixture("get_dns_zone.json"),
		).
		Build(t)

	zone, err := client.GetDNSZone(t.Context(), 1)
	require.NoError(t, err)

	expected := &DNSZone{
		ID:   1,
		Name: "example.com",
		Records: []DNSRecord{
			{Name: "www", Type: "A", Data: "195.230.64.13"},
			{Name: "mail", Type: "MX", Data: "10 mail.client.ru."},
			{Name: "cdn", Type: "A", ConfigRef: &Identifier{ID: 12}},
		},
		Comment: "Комментарий, указанный при последнем редактировании зоны.",
		DNSSec: &DNSSec{
			Enabled:      true,
			DNSSecKeyRef: &Identifier{ID: 1234},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_UpdateDNSZone(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /dns-zone/1",
			servermock.ResponseFromFixture("update_dns_zone.json"),
			servermock.CheckRequestJSONBodyFromFixture("update_dns_zone-request_add.json"),
		).
		Build(t)

	zoneUpdate := DNSZoneUpdate{
		Records: []DNSRecord{
			{Name: "www", Type: "A", Data: "195.230.64.13"},
			{Name: "mail", Type: "MX", Data: "10 mail.client.ru."},
			{Name: "cdn", Type: "A", ConfigRef: &Identifier{ID: 12}},
			{
				Name: "_acme-challenge",
				Type: "TXT",
				Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			},
		},
		Comment: "Комментарий, указанный при последнем редактировании зоны.",
		DNSSec: &DNSSec{
			Enabled:      true,
			DNSSecKeyRef: &Identifier{ID: 1234},
		},
	}

	updated, err := client.UpdateDNSZone(t.Context(), 1, zoneUpdate)
	require.NoError(t, err)

	expected := &DNSZone{
		ID:   1,
		Name: "example.com",
		Records: []DNSRecord{
			{Name: "www", Type: "A", Data: "195.230.64.13"},
			{Name: "mail", Type: "MX", Data: "10 mail.client.ru."},
			{Name: "cdn", Type: "A", Data: "", ConfigRef: &Identifier{ID: 12}},
			{Name: "_acme-challenge", Type: "TXT", Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		},
	}

	assert.Equal(t, expected, updated)
}
