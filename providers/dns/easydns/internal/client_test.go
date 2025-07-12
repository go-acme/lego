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
			client := NewClient("tok", "k")
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("tok", "k"),
	)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/records/all/example.com", servermock.ResponseFromFixture("list-zone.json")).
		Build(t)

	zones, err := client.ListZones(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []ZoneRecord{{
		ID:       "60898922",
		Domain:   "example.com",
		Host:     "hosta",
		TTL:      "300",
		Priority: "0",
		Type:     "A",
		Rdata:    "1.2.3.4",
		LastMod:  "2019-08-28 19:09:50",
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/records/all/example.com", servermock.ResponseFromFixture("error1.json")).
		Build(t)

	_, err := client.ListZones(t.Context(), "example.com")
	require.EqualError(t, err, "code 420: Enhance Your Calm. Rate limit exceeded (too many requests) OR you did NOT provide any credentials with your request!")
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /zones/records/add/example.com/TXT",
			servermock.ResponseFromFixture("add-record.json").WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBody(`{"domain":"example.com","host":"test631","ttl":"300","prio":"0","type":"TXT","rdata":"txt"}`)).
		Build(t)

	record := ZoneRecord{
		Domain:   "example.com",
		Host:     "test631",
		Type:     "TXT",
		Rdata:    "txt",
		TTL:      "300",
		Priority: "0",
	}

	recordID, err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	assert.Equal(t, "xxx", recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /zones/records/add/example.com/TXT",
			servermock.ResponseFromFixture("error1.json").WithStatusCode(http.StatusCreated)).
		Build(t)

	record := ZoneRecord{
		Domain:   "example.com",
		Host:     "test631",
		Type:     "TXT",
		Rdata:    "txt",
		TTL:      "300",
		Priority: "0",
	}

	_, err := client.AddRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "code 420: Enhance Your Calm. Rate limit exceeded (too many requests) OR you did NOT provide any credentials with your request!")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/records/example.com/xxx", nil).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "xxx")
	require.NoError(t, err)
}
