package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(
				WithAuthKey("foo@example.com", "secret"),
				WithHTTPClient(server.Client()),
				WithBaseURL(server.URL),
			)
			if err != nil {
				return nil, err
			}

			return client, nil
		},
		servermock.CheckHeader().
			WithRegexp("User-Agent", `goacme-lego/[0-9.]+ \(.+\)`).
			WithAccept("application/json").
			With("X-Auth-Email", "foo@example.com").
			With("X-Auth-Key", "secret"),
	)
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/023e105f4ecef8ad9ca31a8372d0c353/dns_records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckHeader().
				WithContentType("application/json"),
			servermock.CheckRequestJSONBodyFromFile("create_record-request.json")).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.example.com",
		TTL:     120,
		Type:    "TXT",
		Content: `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`,
	}

	newRecord, err := client.CreateDNSRecord(t.Context(), "023e105f4ecef8ad9ca31a8372d0c353", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "023e105f4ecef8ad9ca31a8372d0c353",
		Name:    "example.com",
		TTL:     3600,
		Type:    "A",
		Comment: "Domain verification record",
		Content: "198.51.100.4",
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_CreateDNSRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/023e105f4ecef8ad9ca31a8372d0c353/dns_records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.example.com",
		TTL:     120,
		Type:    "TXT",
		Content: `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`,
	}

	_, err := client.CreateDNSRecord(t.Context(), "023e105f4ecef8ad9ca31a8372d0c353", record)
	require.EqualError(t, err, "[status code 400] 6003: Invalid request headers; 6103: Invalid format for X-Auth-Key header")
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/023e105f4ecef8ad9ca31a8372d0c353/dns_records/xxx",
			servermock.ResponseFromFixture("delete_record.json")).
		Build(t)

	err := client.DeleteDNSRecord(context.Background(), "023e105f4ecef8ad9ca31a8372d0c353", "xxx")
	require.NoError(t, err)
}

func TestClient_DeleteDNSRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/023e105f4ecef8ad9ca31a8372d0c353/dns_records/xxx",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.DeleteDNSRecord(context.Background(), "023e105f4ecef8ad9ca31a8372d0c353", "xxx")
	require.EqualError(t, err, "[status code 400] 6003: Invalid request headers; 6103: Invalid format for X-Auth-Key header")
}

func TestClient_ZonesByName(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com").
				With("per_page", "50")).
		Build(t)

	zones, err := client.ZonesByName(context.Background(), "example.com")
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:      "023e105f4ecef8ad9ca31a8372d0c353",
			Account: Account{ID: "023e105f4ecef8ad9ca31a8372d0c353", Name: "Example Account Name"},
			Meta: Meta{
				CdnOnly:                true,
				CustomCertificateQuota: 1,
				DNSOnly:                true,
				FoundationDNS:          true,
				PageRuleQuota:          100,
				PhishingDetected:       false,
				Step:                   2,
			},
			Name: "example.com",
			Owner: Owner{
				ID:   "023e105f4ecef8ad9ca31a8372d0c353",
				Name: "Example Org",
				Type: "organization",
			},
			Plan: Plan{
				ID:                "023e105f4ecef8ad9ca31a8372d0c353",
				CanSubscribe:      false,
				Currency:          "USD",
				ExternallyManaged: false,
				Frequency:         "monthly",
				IsSubscribed:      false,
				LegacyDiscount:    false,
				LegacyID:          "free",
				Price:             10,
				Name:              "Example Org",
			},
			CnameSuffix: "cdn.cloudflare.com",
			Paused:      true,
			Permissions: []string{"#worker:read"},
			Tenant: Tenant{
				ID:   "023e105f4ecef8ad9ca31a8372d0c353",
				Name: "Example Account Name",
			},
			TenantUnit: TenantUnit{
				ID: "023e105f4ecef8ad9ca31a8372d0c353",
			},
			Type:              "full",
			VanityNameServers: []string{"ns1.example.com", "ns2.example.com"},
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_ZonesByName_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.ZonesByName(context.Background(), "example.com")
	require.EqualError(t, err, "[status code 400] 6003: Invalid request headers; 6103: Invalid format for X-Auth-Key header")
}
