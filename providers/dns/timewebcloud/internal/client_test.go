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
			client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /v1/domains/example.com/dns-records",
			servermock.ResponseFromFixture("createDomainDNSRecord.json"),
			servermock.CheckRequestJSONBody(`{"type":"TXT","value":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","subdomain":"_acme-challenge"}`)).
		Build(t)

	payload := DNSRecord{
		Type:      "TXT",
		Value:     "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		SubDomain: "_acme-challenge",
	}

	response, err := client.CreateRecord(t.Context(), "example.com.", payload)
	require.NoError(t, err)

	expected := &DNSRecord{
		Type: "TXT",
		ID:   123,
	}

	assert.Equal(t, expected, response)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /v1/domains/example.com/dns-records",
			servermock.ResponseFromFixture("error_bad_request.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.CreateRecord(t.Context(), "example.com.", DNSRecord{})
	require.Error(t, err)

	assert.EqualError(t, err, "400: Value must be a number conforming to the specified constraints (bad_request) [15095f25-aac3-4d60-a788-96cb5136f186]")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/domains/example.com/dns-records/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com.", 123)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/domains/example.com/dns-records/123",
			servermock.ResponseFromFixture("error_unauthorized.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com.", 123)
	require.Error(t, err)

	assert.EqualError(t, err, "401: Unauthorized (unauthorized) [15095f25-aac3-4d60-a788-96cb5136f186]")
}
