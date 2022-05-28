package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()

	server := httptest.NewServer(mux)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dns/example.com/addRR", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get(authenticationHeader) == "" {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		all, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		query, err := url.ParseQuery(string(all))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if query.Get("data") != "txtTXTtxt" {
			http.Error(rw, fmt.Sprintf("data: got %s want txtTXTtxt", query.Get("data")), http.StatusBadRequest)
			return
		}
		if query.Get("name") != "sub" {
			http.Error(rw, fmt.Sprintf("name: got %s want sub", query.Get("name")), http.StatusBadRequest)
			return
		}
		if query.Get("type") != "TXT" {
			http.Error(rw, fmt.Sprintf("type: got %s want TXT", query.Get("TXT")), http.StatusBadRequest)
			return
		}
		if query.Get("ttl") != "30" {
			http.Error(rw, fmt.Sprintf("ttl: got %s want 30", query.Get("ttl")), http.StatusBadRequest)
			return
		}
	})

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord("example.com", record)
	require.NoError(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dns/example.com/removeRR", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get(authenticationHeader) == "" {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		all, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		query, err := url.ParseQuery(string(all))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if query.Get("data") != "txtTXTtxt" {
			http.Error(rw, fmt.Sprintf("data: got %s want txtTXTtxt", query.Get("data")), http.StatusBadRequest)
			return
		}
		if query.Get("name") != "sub" {
			http.Error(rw, fmt.Sprintf("name: got %s want sub", query.Get("name")), http.StatusBadRequest)
			return
		}
		if query.Get("type") != "TXT" {
			http.Error(rw, fmt.Sprintf("type: got %s want TXT", query.Get("TXT")), http.StatusBadRequest)
			return
		}
	})

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord("example.com", record)
	require.NoError(t, err)
}
