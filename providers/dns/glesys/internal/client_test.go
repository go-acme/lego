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

		apiUser, apiKey, ok := req.BasicAuth()
		if apiUser != "user" || apiKey != "secret" || !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/domain/addrecord", http.StatusOK, "add-record.json")

	recordID, err := client.AddTXTRecord(context.Background(), "example.com", "foo", "txt", 120)
	require.NoError(t, err)

	assert.Equal(t, 123, recordID)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/domain/deleterecord", http.StatusOK, "delete-record.json")

	err := client.DeleteTXTRecord(context.Background(), 123)
	require.NoError(t, err)
}
