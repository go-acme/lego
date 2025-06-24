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

		query := req.URL.Query()
		if query.Get("token") != "secret" {
			http.Error(rw, fmt.Sprintf("invalid credentials: %q", query.Get("token")), http.StatusUnauthorized)
			return
		}

		if query.Get("domain") != "example.com" {
			http.Error(rw, fmt.Sprintf("invalid domain: %q", query.Get("domain")), http.StatusBadRequest)
			return
		}

		if query.Has("action") {
			if query.Get("action") != "delete" {
				http.Error(rw, fmt.Sprintf("invalid action: %q", query.Get("action")), http.StatusBadRequest)
				return
			}
		} else {
			if query.Get("value") != "value" {
				http.Error(rw, fmt.Sprintf("invalid value: %q", query.Get("value")), http.StatusBadRequest)
				return
			}
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

	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_UpdateTxtRecord(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/letsencrypt", http.StatusOK, "success.json")

	err := client.UpdateTxtRecord(t.Context(), "example.com.", "value", false)
	require.NoError(t, err)
}

func TestClient_UpdateTxtRecord_clear(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/letsencrypt", http.StatusOK, "success.json")

	err := client.UpdateTxtRecord(t.Context(), "example.com.", "value", true)
	require.NoError(t, err)
}
