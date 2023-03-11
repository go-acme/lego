package internal

import (
	"context"
	"fmt"
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

func setupTest(t *testing.T, method, pattern string, status int, file string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		if file == "" {
			rw.WriteHeader(status)
			return
		}

		open, err := os.Open(filepath.Join("fixtures", file))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.BaseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_GetTxtRecords(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/dns/example.com/txt", http.StatusOK, "get-txt-records.json")

	records, err := client.GetTxtRecords(context.Background(), "example.com")
	require.NoError(t, err)

	expected := []TXTRecord{
		{ID: "123", Name: "prefix.example.com", Destination: "server.example.com", Delete: true, Modify: true, ResourceURL: "string"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddTxtRecord(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/dns/example.com/txt", http.StatusCreated, "create-txt-record.json")

	records, err := client.AddTxtRecord(context.Background(), "example.com", TXTRecord{Name: "prefix.example.com", Destination: "server.example.com"})
	require.NoError(t, err)

	expected := []TXTRecord{
		{ID: "123", Name: "prefix.example.com", Destination: "server.example.com", Delete: true, Modify: true, ResourceURL: "string"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/dns/example.com/txt/123", http.StatusNoContent, "")

	err := client.RemoveTxtRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)
}
