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

const apiKey = "secret"

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(apiKey)
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("api", apiKey),
	)
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains.json", servermock.ResponseFromFixture("get-domains.json")).
		Build(t)

	domains, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{{
		ID:          123,
		UnicodeFqdn: "example.com",
		Domain:      "example",
		TLD:         "com",
		Status:      "ok",
	}}
	assert.Equal(t, expected, domains)
}

func TestClient_GetDomains_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains.json",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.GetDomains(t.Context())
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/123/records.json", servermock.ResponseFromFixture("get-records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), 123)
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      1234,
			Content: "ns1.lima-city.de",
			Name:    "example.com",
			TTL:     36000,
			Type:    "NS",
		},
		{
			ID:      5678,
			Content: `"foobar"`,
			Name:    "_acme-challenge.example.com",
			TTL:     36000,
			Type:    "TXT",
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/123/records.json",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.GetRecords(t.Context(), 123)
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/123/records.json",
			servermock.ResponseFromFixture("ok.json"),
			servermock.CheckRequestJSONBody(`{"nameserver_record":{"name":"foo","content":"bar","ttl":12,"type":"TXT"}}`)).
		Build(t)

	record := Record{
		Name:    "foo",
		Content: "bar",
		TTL:     12,
		Type:    "TXT",
	}

	err := client.AddRecord(t.Context(), 123, record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/123/records.json",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	record := Record{
		Name:    "foo",
		Content: "bar",
		TTL:     12,
		Type:    "TXT",
	}

	err := client.AddRecord(t.Context(), 123, record)
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_UpdateRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /domains/123/records/456",
			servermock.ResponseFromFixture("ok.json"),
			servermock.CheckRequestJSONBody(`{"nameserver_record":{}}`)).
		Build(t)

	err := client.UpdateRecord(t.Context(), 123, 456, Record{})
	require.NoError(t, err)
}

func TestClient_UpdateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /domains/123/records/456",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.UpdateRecord(t.Context(), 123, 456, Record{})
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/123/records/456",
			servermock.ResponseFromFixture("ok.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/123/records/456",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}
