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

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern string, handlerFunc http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, handlerFunc)

	client := NewClient("bob", "user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func authenticatedHandler(method string, status int, file string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		token := req.Header.Get(authTokenHeader)
		if token != "tok" {
			http.Error(rw, fmt.Sprintf("invalid credentials: %q", token), http.StatusUnauthorized)
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
	}
}

func unauthenticatedHandler(method string, status int, file string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		token := req.Header.Get(authTokenHeader)
		if token != "" {
			http.Error(rw, fmt.Sprintf("invalid credentials: %q", token), http.StatusUnauthorized)
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
	}
}

func TestClient_Publish(t *testing.T) {
	client := setupTest(t, "/Zone/example.com", unauthenticatedHandler(http.MethodPut, http.StatusOK, "publish.json"))

	err := client.Publish(t.Context(), "example.com", "my message")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := setupTest(t, "/TXTRecord/example.com/example.com.", unauthenticatedHandler(http.MethodPost, http.StatusCreated, "create-txt-record.json"))

	err := client.AddTXTRecord(t.Context(), "example.com", "example.com.", "txt", 120)
	require.NoError(t, err)
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := setupTest(t, "/TXTRecord/example.com/example.com.", unauthenticatedHandler(http.MethodDelete, http.StatusOK, ""))

	err := client.RemoveTXTRecord(t.Context(), "example.com", "example.com.")
	require.NoError(t, err)
}
