package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, message string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("user", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL = server.URL

	mux.HandleFunc("GET /", func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		username := query.Get("username")
		if username != "user" {
			http.Error(rw, "invalid username: "+username, http.StatusUnauthorized)
			return
		}

		password := query.Get("password")
		if password != "secret" {
			http.Error(rw, "invalid password: "+password, http.StatusUnauthorized)
			return
		}

		_, _ = rw.Write([]byte(message))
	})

	return client
}

func TestAddTXTRecord(t *testing.T) {
	client := setupTest(t, "success")

	err := client.AddTXTRecord(t.Context(), "example.com", "sub.example.com", "value")
	require.NoError(t, err)
}

func TestAddTXTRecord_error(t *testing.T) {
	client := setupTest(t, "error: authentification failed")

	err := client.AddTXTRecord(t.Context(), "example.com", "sub.example.com", "value")
	require.EqualError(t, err, "error: authentification failed")
}
