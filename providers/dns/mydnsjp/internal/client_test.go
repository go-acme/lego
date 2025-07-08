package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, cmdName string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("invalid method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		username, password, ok := req.BasicAuth()
		if !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if username != "xxx" {
			http.Error(rw, fmt.Sprintf("username: want %s got %s", username, "xxx"), http.StatusUnauthorized)
			return
		}

		if password != "secret" {
			http.Error(rw, fmt.Sprintf("password: want %s got %s", password, "secret"), http.StatusUnauthorized)
			return
		}

		if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			http.Error(rw, fmt.Sprintf("invalid Content-Type: %s", req.Header.Get("Content-Type")), http.StatusBadRequest)
			return
		}

		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		domain := req.Form.Get("CERTBOT_DOMAIN")
		if domain != "example.com" {
			http.Error(rw, fmt.Sprintf("unexpected CERTBOT_DOMAIN: %s", domain), http.StatusBadRequest)
			return
		}

		validation := req.Form.Get("CERTBOT_VALIDATION")
		if validation != "txt" {
			http.Error(rw, fmt.Sprintf("unexpected CERTBOT_VALIDATION: %s", validation), http.StatusBadRequest)
			return
		}

		cmd := req.Form.Get("EDIT_CMD")
		if cmd != cmdName {
			http.Error(rw, fmt.Sprintf("unexpected EDIT_CMD: %s", cmd), http.StatusBadRequest)
			return
		}
	})

	client := NewClient("xxx", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := setupTest(t, "REGIST")

	err := client.AddTXTRecord(t.Context(), "example.com", "txt")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := setupTest(t, "DELETE")

	err := client.DeleteTXTRecord(t.Context(), "example.com", "txt")
	require.NoError(t, err)
}
