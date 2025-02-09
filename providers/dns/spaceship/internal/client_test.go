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

	client, err := NewClient("key", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "PUT /dns/records/example.com", http.StatusOK, "")

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "PUT /dns/records/example.com", http.StatusUnprocessableEntity, "error.json")

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.AddRecord(context.Background(), "example.com", record)
	require.EqualError(t, err, "^$, name: The domain name contains invalid characters")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "DELETE /dns/records/example.com", http.StatusOK, "")

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.DeleteRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "DELETE /dns/records/example.com", http.StatusUnprocessableEntity, "error.json")

	record := Record{
		Type: "TXT",
		Name: "@",
		TTL:  60,
	}

	err := client.DeleteRecord(context.Background(), "example.com", record)
	require.EqualError(t, err, "^$, name: The domain name contains invalid characters")
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, "GET /dns/records/example.com", http.StatusOK, "get-records.json")

	records, err := client.GetRecords(context.Background(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{Type: "A", Name: "@", TTL: 3600},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := setupTest(t, "GET /dns/records/example.com", http.StatusUnprocessableEntity, "error.json")

	_, err := client.GetRecords(context.Background(), "example.com")
	require.EqualError(t, err, "^$, name: The domain name contains invalid characters")
}
