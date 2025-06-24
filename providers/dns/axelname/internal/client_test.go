package internal

import (
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

	client, err := NewClient("user", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_ListRecords(t *testing.T) {
	client := setupTest(t, "GET /dns_list", http.StatusOK, "dns_list.json")

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
	client := setupTest(t, "GET /dns_list", http.StatusNotFound, "dns_list_error.json")

	_, err := client.ListRecords(t.Context(), "example.com")
	require.EqualError(t, err, "error: Domain not found (1)")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "GET /dns_delete", http.StatusOK, "dns_delete.json")

	record := Record{ID: "74749"}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "GET /dns_delete", http.StatusNotFound, "dns_delete_error.json")

	record := Record{ID: "74749"}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "error: Domain not found (1)")
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "GET /dns_add", http.StatusOK, "dns_add.json")

	record := Record{ID: "74749"}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "GET /dns_add", http.StatusNotFound, "dns_add_error.json")

	record := Record{ID: "74749"}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "error: Domain not found (1)")
}
