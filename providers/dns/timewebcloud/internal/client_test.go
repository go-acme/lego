package internal

import (
	"net/http"
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
			client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /v2/domains/_acme-challenge.example.com/dns-records",
			servermock.ResponseFromFixture("createDomainDNSRecord.json"),
			servermock.CheckRequestJSONBodyFromFixture("createDomainDNSRecord-request.json"),
		).
		Build(t)

	payload := DNSRecordRequest{
		Type:  "TXT",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	response, err := client.CreateRecord(t.Context(), "_acme-challenge.example.com.", payload)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:   123,
		Type: "TXT",
		Fqdn: "example.com",
		Data: Data{
			Value:     payload.Value,
			Subdomain: "_acme-challenge",
		},
	}

	assert.Equal(t, expected, response)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /v2/domains/_acme-challenge.example.com/dns-records",
			servermock.ResponseFromFixture("error_bad_request.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	_, err := client.CreateRecord(t.Context(), "_acme-challenge.example.com.", DNSRecordRequest{})
	require.Error(t, err)

	assert.EqualError(t, err, "400: Value must be a number conforming to the specified constraints (bad_request) [15095f25-aac3-4d60-a788-96cb5136f186]")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/domains/_acme-challenge.example.com/dns-records/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "_acme-challenge.example.com.", 123)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/domains/_acme-challenge.example.com/dns-records/123",
			servermock.ResponseFromFixture("error_unauthorized.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "_acme-challenge.example.com.", 123)
	require.Error(t, err)

	assert.EqualError(t, err, "401: Unauthorized (unauthorized) [15095f25-aac3-4d60-a788-96cb5136f186]")
}
