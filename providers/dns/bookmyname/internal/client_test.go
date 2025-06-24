package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("GET /", func(rw http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if username != "user" || password != "secret" || !ok {
			http.Error(rw, fmt.Sprintf("auth: user %s, password %s, malformed", username, password), http.StatusOK)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(http.StatusOK)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient("user", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL = server.URL

	return client
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "add_success.txt")

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "error.txt")

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.AddRecord(t.Context(), record)
	require.Error(t, err)

	require.EqualError(t, err, "unexpected response: notfqdn: Host _acme-challenge.sub.example.com. malformed / vhn")
}

func TestClient_RemoveRecord(t *testing.T) {
	client := setupTest(t, "remove_success.txt")

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.RemoveRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := setupTest(t, "error.txt")

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.RemoveRecord(t.Context(), record)
	require.Error(t, err)

	require.EqualError(t, err, "unexpected response: notfqdn: Host _acme-challenge.sub.example.com. malformed / vhn")
}
