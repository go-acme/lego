package internal

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern string, status int, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if filename == "" {
			rw.WriteHeader(status)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient("secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_ListDomains(t *testing.T) {
	client := setupTest(t, "GET /domain", http.StatusOK, "domains.json")

	domains, err := client.ListDomains(context.Background())
	require.NoError(t, err)

	expected := []Domain{
		{ID: 1, Domain: "example.com"},
		{ID: 2, Domain: "example.org"},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_ListDomains_error(t *testing.T) {
	client := setupTest(t, "GET /domain", http.StatusForbidden, "error.json")

	_, err := client.ListDomains(context.Background())
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}

func TestClient_ListRecords(t *testing.T) {
	client := setupTest(t, "GET /domain/123/dns", http.StatusOK, "records.json")

	records, err := client.ListRecords(context.Background(), 123)
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
	client := setupTest(t, "GET /domain/123/dns", http.StatusForbidden, "error.json")

	_, err := client.ListRecords(context.Background(), 123)
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "POST /domain/123/dns", http.StatusOK, "")

	record := Record{
		Host:  "_acme-challenge.foo.example.com",
		Line:  "DEFAULT",
		TTL:   120,
		Type:  "TXT",
		Value: "foo",
	}

	err := client.AddRecord(context.Background(), 123, record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "POST /domain/123/dns", http.StatusForbidden, "error.json")

	record := Record{
		Host:  "_acme-challenge.foo.example.com",
		Line:  "DEFAULT",
		TTL:   120,
		Type:  "TXT",
		Value: "foo",
	}

	err := client.AddRecord(context.Background(), 123, record)
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "DELETE /domain/123/dns", http.StatusOK, "")

	err := client.DeleteRecord(context.Background(), 123, 456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "DELETE /domain/123/dns", http.StatusForbidden, "error.json")

	err := client.DeleteRecord(context.Background(), 123, 456)
	require.Error(t, err)

	assert.EqualError(t, err, "30039: 密钥认证错误或已失效")
}
