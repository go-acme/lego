package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_AddRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/ddns/update.php", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok || username != "anonymous" || password != "secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
		if q.Get("action") != "add" {
			http.Error(w, "invalid action", http.StatusBadRequest)
			return
		}
		if q.Get("zone") != "example.com" {
			http.Error(w, "invalid zone", http.StatusBadRequest)
			return
		}
		if q.Get("type") != "TXT" {
			http.Error(w, "invalid type", http.StatusBadRequest)
			return
		}
		if q.Get("record") != "_acme-challenge.example.com." {
			http.Error(w, "invalid record", http.StatusBadRequest)
			return
		}
		if q.Get("data") != "token" {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	client, err := NewClient(server.URL, "secret", nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.AddRecord(ctx, "example.com", "_acme-challenge.example.com.", "token")
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/ddns/update.php", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok || username != "anonymous" || password != "secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
		if q.Get("action") != "delete" {
			http.Error(w, "invalid action", http.StatusBadRequest)
			return
		}
		if q.Get("zone") != "example.com" {
			http.Error(w, "invalid zone", http.StatusBadRequest)
			return
		}
		if q.Get("type") != "TXT" {
			http.Error(w, "invalid type", http.StatusBadRequest)
			return
		}
		if q.Get("record") != "_acme-challenge.example.com." {
			http.Error(w, "invalid record", http.StatusBadRequest)
			return
		}
		if q.Get("data") != "token" {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	client, err := NewClient(server.URL, "secret", nil)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.DeleteRecord(ctx, "example.com", "_acme-challenge.example.com.", "token")
	require.NoError(t, err)
}
