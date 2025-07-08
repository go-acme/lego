package internal

import (
	"fmt"
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
	t.Cleanup(server.Close)

	client := NewClient("user", "secret")
	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client.baseURL = serverURL.String()

	return client, mux
}

func TestClient_UpdateTXTRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("GET /", func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		fields := map[string]string{
			"username": "user",
			"password": "secret",
			"rid":      "123456",
			"content":  "txt",
		}

		for k, v := range fields {
			value := query.Get(k)
			if value != v {
				http.Error(rw, fmt.Sprintf("%s: unexpected value: %s (%s)", k, value, v), http.StatusBadRequest)
				return
			}
		}
	})

	err := client.UpdateTXTRecord(t.Context(), "123456", "txt")
	require.NoError(t, err)
}

func TestClient_UpdateTXTRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("GET /", func(rw http.ResponseWriter, _ *http.Request) {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	})

	err := client.UpdateTXTRecord(t.Context(), "123456", "txt")
	require.Error(t, err)
}
