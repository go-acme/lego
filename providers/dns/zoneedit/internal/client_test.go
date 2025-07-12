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

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("user", "secret")
	require.NoError(t, err)

	client.baseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, mux
}

func fixtureHandler(filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		if filename == "" {
			return
		}

		open, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/txt-create.php", fixtureHandler("success.xml"))

	err := client.CreateTXTRecord("_acme-challenge.example.com", "value")
	require.NoError(t, err)
}

func TestClient_CreateTXTRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/txt-create.php", fixtureHandler("error.xml"))

	err := client.CreateTXTRecord("_acme-challenge.example.com", "value")
	require.EqualError(t, err, "[status code: 200] 708: Failed Login: user (_acme-challenge.example.com)")
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/txt-delete.php", fixtureHandler("success.xml"))

	err := client.DeleteTXTRecord("_acme-challenge.example.com", "value")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/txt-delete.php", fixtureHandler("error.xml"))

	err := client.DeleteTXTRecord("_acme-challenge.example.com", "value")
	require.EqualError(t, err, "[status code: 200] 708: Failed Login: user (_acme-challenge.example.com)")
}
