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
		Route("GET /dns/zones",
			servermock.ResponseFromFixture("zones.json")).
		Build(t)

	zones, err := client.GetZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{ServerID: "25", AccountID: "6894", DomainID: "9149", Domain: "example.com", DisplayName: "example.com"},
		{ServerID: "25", AccountID: "6894", DomainID: "9150", Domain: "example.org", DisplayName: "example.org"},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZones(t.Context())
	require.EqualError(t, err, "401: Authentication required (UNAUTHORIZED) [requestId=00000000-0000-0000-0000-000000000000]")
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/zones/9149/records",
			servermock.ResponseFromFixture("record.json")).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "_acme-challenge",
		Value: "txt-value",
		TTL:   60,
	}

	newRecord, err := client.AddRecord(t.Context(), "9149", record)
	require.NoError(t, err)

	expected := &Record{
		ID:     "drr_06ezwatrgahtygnvpz8cp995y0",
		Type:   "TXT",
		Name:   "_acme-challenge.example.com",
		Value:  `"txt-value"`,
		TTL:    60,
		Status: "pending",
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/zones/9149/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "_acme-challenge",
		Value: "txt-value",
		TTL:   60,
	}

	_, err := client.AddRecord(t.Context(), "9149", record)
	require.EqualError(t, err, "401: Authentication required (UNAUTHORIZED) [requestId=00000000-0000-0000-0000-000000000000]")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/zones/9149/records/drr_06ezwatrgahtygnvpz8cp995y0",
			servermock.Noop().WithStatusCode(http.StatusOK)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "9149", "drr_06ezwatrgahtygnvpz8cp995y0")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/zones/9149/records/drr_06ezwatrgahtygnvpz8cp995y0",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "9149", "drr_06ezwatrgahtygnvpz8cp995y0")
	require.EqualError(t, err, "401: Authentication required (UNAUTHORIZED) [requestId=00000000-0000-0000-0000-000000000000]")
}
