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
			opts := Options{
				Login:    "l",
				Username: "u",
				Password: "p",
				OTP:      "2hsn",
			}

			client := NewClient(opts)
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("l.u", "p").
			WithRegexp(headerTOTPToken, `\d{6}`))
}

func TestClient_GetZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /anycast/nicmanager-anycastdns4.net",
			servermock.ResponseFromFixture("zone.json")).
		Build(t)

	zone, err := client.GetZone(t.Context(), "nicmanager-anycastdns4.net")
	require.NoError(t, err)

	expected := &Zone{
		Name:   "nicmanager-anycastdns4.net",
		Active: true,
		Records: []Record{
			{
				ID:      186,
				Name:    "nicmanager-anycastdns4.net",
				Type:    "A",
				Content: "123.123.123.123",
				TTL:     3600,
			},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_GetZone_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /anycast/foo",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	_, err := client.GetZone(t.Context(), "foo")
	require.EqualError(t, err, "404: Not Found")
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /anycast/zonedomain.tld/records",
			servermock.Noop().
				WithStatusCode(http.StatusAccepted)).
		Build(t)

	record := RecordCreateUpdate{
		Type:  "TXT",
		Name:  "lego",
		Value: "content",
		TTL:   3600,
	}

	err := client.AddRecord(t.Context(), "zonedomain.tld", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /anycast/zonedomain.tld/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := RecordCreateUpdate{
		Type:  "TXT",
		Name:  "zonedomain.tld",
		Value: "content",
		TTL:   3600,
	}

	err := client.AddRecord(t.Context(), "zonedomain.tld", record)
	require.EqualError(t, err, "401: Not Found")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /anycast/zonedomain.tld/records/6",
			servermock.Noop().
				WithStatusCode(http.StatusAccepted)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "zonedomain.tld", 6)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /anycast/zonedomain.tld/records/6",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "zonedomain.tld", 6)
	require.EqualError(t, err, "404: Not Found")
}
