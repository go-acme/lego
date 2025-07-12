package internal

import (
	"context"
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
			client := NewClient(context.Background(), "STACK_ID", "CLIENT_ID", "CLIENT_SECRET")
			client.httpClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL + "/")

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_GetZoneRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /STACK_ID/zones/A/records",
			servermock.ResponseFromFixture("get_zone_records.json"),
			servermock.CheckQueryParameter().Strict().
				With("page_request.filter", "name='foo1' and type='TXT'")).
		Build(t)

	records, err := client.GetZoneRecords(t.Context(), "foo1", &Zone{ID: "A", Domain: "test"})
	require.NoError(t, err)

	expected := []Record{
		{ID: "1", Name: "foo1", Type: "TXT", TTL: 120, Data: "txtTXTtxt"},
		{ID: "2", Name: "foo2", Type: "TXT", TTL: 121, Data: "TXTtxtTXT"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetZoneRecords_apiError(t *testing.T) {
	client := mockBuilder().
		Route("GET /STACK_ID/zones/A/records",
			servermock.RawStringResponse(`
{
	"code": 401,
	"error": "an unauthorized request is attempted."
}`).WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZoneRecords(t.Context(), "foo1", &Zone{ID: "A", Domain: "test"})

	expected := &ErrorResponse{Code: 401, Message: "an unauthorized request is attempted."}
	assert.Equal(t, expected, err)
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /STACK_ID/zones",
			servermock.ResponseFromFixture("get_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("page_request.filter", "domain='foo.com'")).
		Build(t)

	zone, err := client.GetZones(t.Context(), "sub.foo.com")
	require.NoError(t, err)

	expected := &Zone{ID: "A", Domain: "foo.com"}

	assert.Equal(t, expected, zone)
}
