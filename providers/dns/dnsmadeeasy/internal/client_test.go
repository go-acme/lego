package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("key", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With("x-dnsme-apiKey", "key").
			WithRegexp("x-dnsme-requestDate", `\w+, \d+ \w+ \d+ \d+:\d+:\d+ UTC`).
			WithRegexp("x-dnsme-hmac", `[a-z0-9]+`),
	)
}

func TestClient_GetDomain(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/managed/name",
			servermock.RawStringResponse(`{"id": 1, "name": "foo"}`),
			servermock.CheckQueryParameter().Strict().
				With("domainname", "example.com")).
		Build(t)

	domain, err := client.GetDomain(t.Context(), "example.com.")
	require.NoError(t, err)

	expected := &Domain{
		ID:   1,
		Name: "foo",
	}

	assert.Equal(t, expected, domain)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/managed/1/records",
			servermock.ResponseFromFixture("get_records.json"),
			servermock.CheckQueryParameter().Strict().
				With("recordName", "foo").
				With("type", "TXT"),
		).
		Build(t)

	domain := &Domain{ID: 1, Name: "foo"}

	records, err := client.GetRecords(t.Context(), domain, "foo", "TXT")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:       1,
			Type:     "TXT",
			Name:     "foo",
			Value:    "aaa",
			TTL:      60,
			SourceID: 123,
		},
		{
			ID:       2,
			Type:     "TXT",
			Name:     "bar",
			Value:    "bbb",
			TTL:      120,
			SourceID: 456,
		},
	}

	assert.Equal(t, &expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/managed/1/records", nil,
			servermock.CheckRequestJSONBodyFromFile("create_record-request.json")).
		Build(t)

	domain := &Domain{ID: 1, Name: "foo"}

	record := &Record{
		ID:       1,
		Type:     "TXT",
		Name:     "foo",
		Value:    "aaa",
		TTL:      60,
		SourceID: 123,
	}

	err := client.CreateRecord(t.Context(), domain, record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/managed/123/records/1", nil).
		Build(t)

	record := Record{
		ID:       1,
		Type:     "TXT",
		Name:     "foo",
		Value:    "aaa",
		TTL:      60,
		SourceID: 123,
	}

	err := client.DeleteRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_sign(t *testing.T) {
	apiKey := "key"

	client := Client{apiKey: apiKey, apiSecret: "secret"}

	req, err := http.NewRequest(http.MethodGet, "", http.NoBody)
	require.NoError(t, err)

	timestamp := time.Date(2015, time.June, 2, 2, 36, 7, 0, time.UTC).Format(time.RFC1123)

	err = client.sign(req, timestamp)
	require.NoError(t, err)

	assert.Equal(t, apiKey, req.Header.Get("x-dnsme-apiKey"))
	assert.Equal(t, timestamp, req.Header.Get("x-dnsme-requestDate"))
	assert.Equal(t, "6b6c8432119c31e1d3776eb4cd3abd92fae4a71c", req.Header.Get("x-dnsme-hmac"))
}
