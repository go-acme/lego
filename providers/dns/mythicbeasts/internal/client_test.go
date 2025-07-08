package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern string, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, handler)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.APIEndpoint, _ = url.Parse(server.URL)
	client.token = &Token{
		Token:     "secret",
		Lifetime:  60,
		TokenType: "bearer",
		Deadline:  time.Now().Add(1 * time.Minute),
	}

	return client
}

func writeFixtureHandler(method, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, _ = io.Copy(rw, file)
	}
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client := setupTest(t, "/zones/example.com/records/foo/TXT", writeFixtureHandler(http.MethodPost, "post-zoneszonerecords.json"))

	err := client.CreateTXTRecord(mockContext(t), "example.com", "foo", "txt", 120)
	require.NoError(t, err)
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := setupTest(t, "/zones/example.com/records/foo/TXT", writeFixtureHandler(http.MethodDelete, "delete-zoneszonerecords.json"))

	err := client.RemoveTXTRecord(mockContext(t), "example.com", "foo", "txt")
	require.NoError(t, err)
}
