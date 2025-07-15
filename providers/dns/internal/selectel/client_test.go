package selectel

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("token")
	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_ListRecords(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders()).
		Route("GET /123/records/", servermock.ResponseFromFixture("list_records.json")).
		Build(t)

	records, err := client.ListRecords(t.Context(), 123)
	require.NoError(t, err)

	expected := []Record{
		{ID: 123, Name: "example.com", Type: "TXT", TTL: 60, Email: "email@example.com", Content: "txttxttxtA"},
		{ID: 1234, Name: "example.org", Type: "TXT", TTL: 60, Email: "email@example.org", Content: "txttxttxtB"},
		{ID: 12345, Name: "example.net", Type: "TXT", TTL: 60, Email: "email@example.net", Content: "txttxttxtC"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			With(tokenHeader, "token")).
		Route("GET /123/records/",
			servermock.ResponseFromFixture("error.json").WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	records, err := client.ListRecords(t.Context(), 123)

	require.EqualError(t, err, "request failed with status code 401: API error: 400 - error description - field that the error occurred in")
	assert.Nil(t, records)
}

func TestClient_GetDomainByName(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			With(tokenHeader, "token")).
		Route("GET /sub.sub.example.org",
			servermock.Noop().WithStatusCode(http.StatusNotFound)).
		Route("GET /sub.example.org",
			servermock.Noop().WithStatusCode(http.StatusNotFound)).
		Route("GET /example.org",
			servermock.ResponseFromFixture("domains.json")).
		Build(t)

	domain, err := client.GetDomainByName(t.Context(), "sub.sub.example.org")
	require.NoError(t, err)

	expected := &Domain{
		ID:   123,
		Name: "example.org",
	}

	assert.Equal(t, expected, domain)
}

func TestClient_AddRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			With(tokenHeader, "token")).
		Route("POST /123/records/",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json")).
		Build(t)

	record, err := client.AddRecord(t.Context(), 123, Record{
		Name:    "example.org",
		Type:    "TXT",
		TTL:     60,
		Email:   "email@example.org",
		Content: "txttxttxttxt",
	})

	require.NoError(t, err)

	expected := &Record{
		ID:      456,
		Name:    "example.org",
		Type:    "TXT",
		TTL:     60,
		Email:   "email@example.org",
		Content: "txttxttxttxt",
	}

	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			With(tokenHeader, "token")).
		Route("DELETE /123/records/456", nil).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.NoError(t, err)
}
