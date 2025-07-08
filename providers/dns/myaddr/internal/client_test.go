package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

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

	credentials := map[string]string{
		"example": "secret",
	}

	client, err := NewClient(credentials)
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := setupTest(t, "POST /update", http.StatusOK, "")

	err := client.AddTXTRecord(t.Context(), "example", "txt")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := setupTest(t, "POST /update", http.StatusBadRequest, "error.txt")

	err := client.AddTXTRecord(t.Context(), "example", "txt")
	require.EqualError(t, err, `unexpected status code: [status code: 400] body: invalid value for "key"`)
}

func TestClient_AddTXTRecord_error_credentials(t *testing.T) {
	client := setupTest(t, "POST /update", http.StatusOK, "")

	err := client.AddTXTRecord(t.Context(), "nx", "txt")
	require.EqualError(t, err, "subdomain nx not found in credentials, check your credentials map")
}
