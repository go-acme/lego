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
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		token, key, ok := req.BasicAuth()
		if token != "tok" || key != "k" || !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if req.URL.Query().Get("format") != "json" {
			http.Error(rw, fmt.Sprintf("invalid format: %s", req.URL.Query().Get("format")), http.StatusBadRequest)
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

	client := NewClient("tok", "k")
	client.HTTPClient = server.Client()
	client.BaseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, http.MethodPut, "/zones/records/add/example.com/TXT", http.StatusCreated, "add-record.json")

	record := ZoneRecord{
		Domain:   "example.com",
		Host:     "test631",
		Type:     "TXT",
		Rdata:    "txt",
		TTL:      "300",
		Priority: "0",
	}

	recordID, err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)

	assert.Equal(t, "xxx", recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/zones/records/example.com/xxx", http.StatusOK, "")

	err := client.DeleteRecord(context.Background(), "example.com", "xxx")
	require.NoError(t, err)
}
