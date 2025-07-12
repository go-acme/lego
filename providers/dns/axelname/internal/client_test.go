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

func setupClient(server *httptest.Server) (*Client, error) {
	client, err := NewClient("user", "secret")
	if err != nil {
		return nil, err
	}

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, nil
}

func TestClient_ListRecords(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /dns_list",
			servermock.ResponseFromFixture("dns_list.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain", "example.com").
				With("nichdl", "user").
				With("token", "secret")).
		Build(t)

	records, err := client.ListRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{ID: "74749", Name: "example.com", Type: "A", Value: "46.161.54.22"},
		{ID: "417", Name: "example.com", Type: "MX", Value: "mx.yandex.ru.", Prio: "10"},
		{ID: "419", Name: "mail.example.com", Type: "CNAME", Value: "mail.yandex.ru."},
		{ID: "74750", Name: "www.example.com", Type: "A", Value: "46.161.54.22"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /dns_list",
			servermock.ResponseFromFixture("dns_list_error.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	_, err := client.ListRecords(t.Context(), "example.com")
	require.EqualError(t, err, "error: Domain not found (1)")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /dns_delete",
			servermock.ResponseFromFixture("dns_delete.json"),
			servermock.CheckQueryParameter().Strict().
				With("id", "74749").
				With("domain", "example.com").
				With("nichdl", "user").
				With("token", "secret")).
		Build(t)

	record := Record{ID: "74749"}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /dns_delete",
			servermock.ResponseFromFixture("dns_delete_error.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	record := Record{ID: "74749"}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "error: Domain not found (1)")
}

func TestClient_AddRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /dns_add",
			servermock.ResponseFromFixture("dns_add.json"),
			servermock.CheckQueryParameter().Strict().
				With("id", "74749").
				With("domain", "example.com").
				With("nichdl", "user").
				With("token", "secret")).
		Build(t)

	record := Record{ID: "74749"}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /dns_add",
			servermock.ResponseFromFixture("dns_add_error.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	record := Record{ID: "74749"}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "error: Domain not found (1)")
}
