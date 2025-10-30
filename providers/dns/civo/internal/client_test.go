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
			client, err := NewClient(OAuthStaticAccessToken(server.Client(), "secret"), "LON1")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("Authorization", "Bearer secret").
			WithRegexp("User-Agent", `goacme-lego/[0-9.]+ \(.+\)`),
	)
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns",
			servermock.ResponseFromFixture("list_domain_names.json"),
			servermock.CheckQueryParameter().Strict().
				With("region", "LON1")).
		Build(t)

	domains, err := client.ListDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{{
		ID:        "7088fcea-7658-43e6-97fa-273f901978fd",
		AccountID: "e7e8386e-434e-482f-95e0-c406e5d564c2",
		Name:      "example.com",
	}}

	assert.Equal(t, expected, domains)
}

func TestClient_ListDNSRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/7088fcea-7658-43e6-97fa-273f901978fd/records",
			servermock.ResponseFromFixture("list_dns_records.json"),
			servermock.CheckQueryParameter().Strict().
				With("region", "LON1")).
		Build(t)

	records, err := client.ListDNSRecords(t.Context(), "7088fcea-7658-43e6-97fa-273f901978fd")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:       "76cc107f-fbef-4e2b-b97f-f5d34f4075d3",
			DomainID: "edc5dacf-a2ad-4757-41ee-c12f06259c70",
			Name:     "_acme-challenge",
			Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			Type:     "txt",
			TTL:      600,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListDNSRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/7088fcea-7658-43e6-97fa-273f901978fd/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.ListDNSRecords(t.Context(), "7088fcea-7658-43e6-97fa-273f901978fd")
	require.EqualError(t, err, "database_account_not_found: Failed to find the account within the internal database")
}

func TestClient_ListDNSRecords_error_raw(t *testing.T) {
	// the API says:
	// > 4xx/5xx status may not be JSON, unless it's obvious that the response should be parsed for a specific reason.
	// > So, for example, 404 Not Found pages are a standard page of text
	// > but 403 Unauthorized requests may have a reason attribute available in the JSON object.
	// https://www.civo.com/api#parameters-and-responses
	client := mockBuilder().
		Route("GET /dns/7088fcea-7658-43e6-97fa-273f901978fd/records",
			servermock.RawStringResponse(http.StatusText(http.StatusNotFound)).
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	_, err := client.ListDNSRecords(t.Context(), "7088fcea-7658-43e6-97fa-273f901978fd")
	require.EqualError(t, err, "unexpected status code: [status code: 404] body: Not Found")
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/7088fcea-7658-43e6-97fa-273f901978fd/records",
			servermock.ResponseFromFixture("create_dns_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_dns_record-request.json")).
		Build(t)

	record := Record{
		Name:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Type:  "TXT",
		TTL:   600,
	}

	newRecord, err := client.CreateDNSRecord(t.Context(), "7088fcea-7658-43e6-97fa-273f901978fd", record)
	require.NoError(t, err)

	expected := &Record{
		ID:       "76cc107f-fbef-4e2b-b97f-f5d34f4075d3",
		DomainID: "edc5dacf-a2ad-4757-41ee-c12f06259c70",
		Name:     "_acme-challenge",
		Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Type:     "txt",
		TTL:      600,
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/edc5dacf-a2ad-4757-41ee-c12f06259c70/records/76cc107f-fbef-4e2b-b97f-f5d34f4075d3",
			servermock.ResponseFromFixture("delete_dns_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("region", "LON1")).
		Build(t)

	record := Record{
		ID:       "76cc107f-fbef-4e2b-b97f-f5d34f4075d3",
		DomainID: "edc5dacf-a2ad-4757-41ee-c12f06259c70",
		Name:     "_acme-challenge",
		Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Type:     "TXT",
		TTL:      600,
	}

	err := client.DeleteDNSRecord(t.Context(), record)
	require.NoError(t, err)
}
