package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func checkParameter(query url.Values, key, expected string) error {
	if query.Get(key) != expected {
		return fmt.Errorf("%s: want %s got %s", key, expected, query.Get(key))
	}
	return nil
}

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func TestClient_Do(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		username, password, ok := req.BasicAuth()
		if !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if username != "user" {
			http.Error(rw, fmt.Sprintf("username: want %s got %s", username, "user"), http.StatusUnauthorized)
			return
		}

		if password != "secret" {
			http.Error(rw, fmt.Sprintf("password: want %s got %s", password, "secret"), http.StatusUnauthorized)
			return
		}

		query := req.URL.Query()

		values := map[string]string{
			"hostname": "example.com",
			"type":     "TXT",
			"value":    "data",
			"ttl":      "120",
		}

		for k, v := range values {
			err := checkParameter(query, k, v)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}
	})

	record := Record{
		Hostname: "example.com",
		Type:     "TXT",
		Value:    "data",
		TTL:      120,
	}

	err := client.SendRequest(t.Context(), record)
	require.NoError(t, err)
}
