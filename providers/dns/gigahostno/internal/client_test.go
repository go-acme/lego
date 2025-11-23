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

	expected := []Zone{{
		ZoneID:           "123",
		CustomerID:       "1111",
		ZoneName:         "example.com",
		ZoneNameDisplay:  "example.com",
		ZoneType:         "NATIVE",
		ZoneActive:       "1",
		ZoneProtected:    "1",
		ZoneIsRegistered: "1",
		DomainRegistrar:  "norid",
		DomainStatus:     "active",
		DomainExpiryDate: "2025-12-31 23:59:59",
		DomainAutoRenew:  "1",
		ExternalDNS:      "0",
		RecordCount:      5,
		ZoneUpdated:      1700000000,
	}}

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
			RecordID:    "abc123",
			RecordName:  "@",
			RecordType:  "A",
			RecordValue: "185.125.168.166",
			RecordTTL:   3600,
		},
		{
			RecordID:    "def456",
			RecordName:  "www",
			RecordType:  "A",
			RecordValue: "185.125.168.166",
			RecordTTL:   3600,
		},
		{
			RecordID:    "ghi789",
			RecordName:  "@",
			RecordType:  "MX",
			RecordValue: "mail.example.no",
			RecordTTL:   3600,
		},
		{
			RecordID:    "jkl012",
			RecordName:  "_acme-challenge",
			RecordType:  "TXT",
			RecordValue: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			RecordTTL:   120,
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
		RecordName:  "_acme-challenge",
		RecordType:  "TXT",
		RecordValue: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		RecordTTL:   120,
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
