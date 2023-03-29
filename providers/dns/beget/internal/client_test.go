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

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/dns/changeRecords", handler)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func writeFixtureHandler(filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		query := req.URL.Query()

		if query.Get("login") != "user" {
			http.Error(rw, fmt.Sprintf("invalid login: %q", query.Get("login")), http.StatusUnauthorized)
			return
		}
		if query.Get("passwd") != "secret" {
			http.Error(rw, fmt.Sprintf("invalid password: %q", query.Get("passwd")), http.StatusUnauthorized)
			return
		}

		if filename == "" {
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

func TestClient_AddTXTRecord(t *testing.T) {
	client := setupTest(t, writeFixtureHandler("changeRecords.json"))

	err := client.AddTXTRecord(context.Background(), "example.com", "sub", "txtTXTtxt")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := setupTest(t, writeFixtureHandler("error.json"))

	err := client.AddTXTRecord(context.Background(), "example.com", "sub", "txtTXTtxt")
	require.Error(t, err)
}

func TestClient_AddTXTRecord_answer_error(t *testing.T) {
	client := setupTest(t, writeFixtureHandler("answer_error.json"))

	err := client.AddTXTRecord(context.Background(), "example.com", "sub", "txtTXTtxt")
	require.Error(t, err)
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	client := setupTest(t, writeFixtureHandler("changeRecords.json"))

	err := client.RemoveTxtRecord(context.Background(), "example.com", "sub")
	require.NoError(t, err)
}

func TestClient_RemoveTxtRecord_error(t *testing.T) {
	client := setupTest(t, writeFixtureHandler("error.json"))

	err := client.RemoveTxtRecord(context.Background(), "example.com", "sub")
	require.Error(t, err)
}
