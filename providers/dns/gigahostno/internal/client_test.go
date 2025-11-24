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
			client := NewClient()

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones",
			servermock.ResponseFromFixture("zones.json")).
		Build(t)

	zones, err := client.GetZones(mockContext(t))
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:               "123",
			Name:             "example.com",
			NameDisplay:      "example.com",
			Type:             "NATIVE",
			Active:           "1",
			Protected:        "1",
			IsRegistered:     "1",
			Updated:          false,
			CustomerID:       "16030",
			DomainRegistrar:  "norid",
			DomainStatus:     "active",
			DomainExpiryDate: "2026-11-23 15:17:38",
			DomainAutoRenew:  "1",
			ExternalDNS:      "0",
			RecordCount:      4,
		},
		{
			ID:               "226",
			Name:             "example.org",
			NameDisplay:      "example.org",
			Type:             "NATIVE",
			Active:           "1",
			Protected:        "1",
			IsRegistered:     "1",
			Updated:          false,
			CustomerID:       "16030",
			DomainRegistrar:  "norid",
			DomainStatus:     "active",
			DomainExpiryDate: "2026-11-23 14:15:01",
			DomainAutoRenew:  "1",
			ExternalDNS:      "0",
			RecordCount:      5,
		},
		{
			ID:               "229",
			Name:             "example.xn--zckzah",
			NameDisplay:      "example.テスト",
			Type:             "NATIVE",
			Active:           "1",
			Protected:        "1",
			IsRegistered:     "1",
			Updated:          false,
			CustomerID:       "16030",
			DomainRegistrar:  "norid",
			DomainStatus:     "active",
			DomainExpiryDate: "2026-12-01 12:40:48",
			DomainAutoRenew:  "1",
			ExternalDNS:      "0",
			RecordCount:      4,
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZones(mockContext(t))
	require.EqualError(t, err, "401: 401 Unauthorized: 401 Unauthorized")
}

func TestClient_GetZoneRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones/123/records",
			servermock.ResponseFromFixture("zone_records.json")).
		Build(t)

	zones, err := client.GetZoneRecords(mockContext(t), "123")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:    "abc123",
			Name:  "@",
			Type:  "A",
			Value: "185.125.168.166",
			TTL:   3600,
		},
		{
			ID:    "def456",
			Name:  "www",
			Type:  "A",
			Value: "185.125.168.166",
			TTL:   3600,
		},
		{
			ID:    "ghi789",
			Name:  "@",
			Type:  "MX",
			Value: "mail.example.no",
			TTL:   3600,
		},
		{
			ID:    "jkl012",
			Name:  "_acme-challenge",
			Type:  "TXT",
			Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			TTL:   120,
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_CreateNewRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/zones/example.com/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := Record{
		Name:  "_acme-challenge",
		Type:  "TXT",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   120,
	}

	err := client.CreateNewRecord(mockContext(t), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("/dns/zones/123/records/abc123",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT")).
		Build(t)

	err := client.DeleteRecord(mockContext(t), "123", "abc123", "_acme-challenge", "TXT")
	require.NoError(t, err)
}
