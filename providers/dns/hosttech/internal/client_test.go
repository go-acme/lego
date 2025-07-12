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

const testAPIKey = "secret"

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(OAuthStaticAccessToken(server.Client(), testAPIKey))
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer secret"))
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /user/v1/zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("limit", "100").
				With("query", "")).
		Build(t)

	zones, err := client.GetZones(t.Context(), "", 100, 0)
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:          10,
			Name:        "user1.ch",
			Email:       "test@hosttech.ch",
			TTL:         10800,
			Nameserver:  "ns1.hosttech.ch",
			Dnssec:      false,
			DnssecEmail: "test@hosttech.ch",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /user/v1/zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZones(t.Context(), "", 100, 0)
	require.Error(t, err)
}

func TestClient_GetZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /user/v1/zones/123",
			servermock.ResponseFromFixture("zone.json")).
		Build(t)

	zone, err := client.GetZone(t.Context(), "123")
	require.NoError(t, err)

	expected := &Zone{
		ID:          10,
		Name:        "user1.ch",
		Email:       "test@hosttech.ch",
		TTL:         10800,
		Nameserver:  "ns1.hosttech.ch",
		Dnssec:      false,
		DnssecEmail: "test@hosttech.ch",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_GetZone_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /user/v1/zones/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZone(t.Context(), "123")
	require.EqualError(t, err, "401: Unauthenticated.")
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /user/v1/zones/123/records",
			servermock.ResponseFromFixture("records.json"),
			servermock.CheckQueryParameter().Strict().
				With("type", "TXT")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "123", "TXT")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      10,
			Type:    "A",
			Name:    "www",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      11,
			Type:    "AAAA",
			Name:    "www",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      12,
			Type:    "CAA",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      13,
			Type:    "CNAME",
			Name:    "www",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      14,
			Type:    "MX",
			Name:    "mail.example.com",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      14,
			Type:    "NS",
			Name:    "ns1.example.com",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      15,
			Type:    "PTR",
			Name:    "smtp.example.com",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      16,
			Type:    "SRV",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      17,
			Type:    "TXT",
			Text:    "v=spf1 ip4:1.2.3.4/32 -all",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      17,
			Type:    "TLSA",
			Text:    "0 0 1 d2abde240d7cd3ee6b4b28c54df034b97983a1d16e8a410e4561cb106618e971",
			TTL:     3600,
			Comment: "my first record",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /user/v1/zones/123/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetRecords(t.Context(), "123", "TXT")
	require.EqualError(t, err, "401: Unauthenticated.")
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /user/v1/zones/123/records",
			servermock.ResponseFromFixture("record.json").
				WithStatusCode(http.StatusCreated)).
		Build(t)

	record := Record{
		Type:    "TXT",
		Name:    "lego",
		Text:    "content",
		TTL:     3600,
		Comment: "example",
	}

	newRecord, err := client.AddRecord(t.Context(), "123", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      10,
		Type:    "TXT",
		Name:    "lego",
		Text:    "content",
		TTL:     3600,
		Comment: "example",
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /user/v1/zones/123/records",
			servermock.ResponseFromFixture("error-details.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Type:    "TXT",
		Name:    "lego",
		Text:    "content",
		TTL:     3600,
		Comment: "example",
	}

	_, err := client.AddRecord(t.Context(), "123", record)
	require.EqualError(t, err, "401: The given data was invalid. type: [Darf nicht leer sein.]")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /user/v1/zones/123/records/6",
			servermock.Noop().WithStatusCode(http.StatusNoContent).
				WithStatusCode(http.StatusCreated)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "123", "6")
	require.Error(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /user/v1/zones/123/records/6",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "123", "6")
	require.EqualError(t, err, "401: Unauthenticated.")
}
