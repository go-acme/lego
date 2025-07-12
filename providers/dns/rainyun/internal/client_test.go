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
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders())
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain",
			servermock.ResponseFromFixture("domains.json"),
			servermock.CheckQueryParameter().Strict().
				With("options", `{"columnFilters":{"domains.Domain":""},"sort":[],"page":1,"perPage":100}`)).
		Build(t)

	domains, err := client.ListDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{
		{ID: 1, Domain: "example.com"},
		{ID: 2, Domain: "example.org"},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_ListDomains_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	_, err := client.ListDomains(t.Context())
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}

func TestClient_ListRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/123/dns",
			servermock.ResponseFromFixture("records.json"),
			servermock.CheckQueryParameter().Strict().
				With("limit", "100").
				With("page_no", "1")).
		Build(t)

	records, err := client.ListRecords(t.Context(), 123)
	require.NoError(t, err)

	expected := []Record{
		{
			ID:    1,
			Host:  "_acme-challenge.foo.example.com",
			Line:  "DEFAULT",
			TTL:   120,
			Type:  "TXT",
			Value: "foo",
		},
		{
			ID:    2,
			Host:  "_acme-challenge.bar.example.com",
			Line:  "DEFAULT",
			TTL:   300,
			Type:  "TXT",
			Value: "bar",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/123/dns",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	_, err := client.ListRecords(t.Context(), 123)
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/123/dns", nil).
		Build(t)

	record := Record{
		Host:  "_acme-challenge.foo.example.com",
		Line:  "DEFAULT",
		TTL:   120,
		Type:  "TXT",
		Value: "foo",
	}

	err := client.AddRecord(t.Context(), 123, record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/123/dns",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	record := Record{
		Host:  "_acme-challenge.foo.example.com",
		Line:  "DEFAULT",
		TTL:   120,
		Type:  "TXT",
		Value: "foo",
	}

	err := client.AddRecord(t.Context(), 123, record)
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domain/123/dns", nil).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domain/123/dns",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}
