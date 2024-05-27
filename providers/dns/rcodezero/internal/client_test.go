package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

		apiToken := req.Header.Get(authorizationHeader)
		if apiToken != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid credentials: %s", apiToken), http.StatusBadRequest)
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
	})

	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_UpdateRecords_error(t *testing.T) {
	client := setupTest(t, http.MethodPatch, "/v1/acme/zones/example.org/rrsets", http.StatusUnprocessableEntity, "error.json")

	rrSet := []UpdateRRSet{{
		Name:       "acme.example.org.",
		ChangeType: "add",
		Type:       "TXT",
		Records:    []Record{{Content: `"my-acme-challenge"`}},
	}}

	resp, err := client.UpdateRecords(context.Background(), "example.org", rrSet)
	require.ErrorAs(t, err, new(*APIResponse))
	assert.Nil(t, resp) //nolint:testifylint // false positive https://github.com/Antonboom/testifylint/issues/95
}

func TestClient_UpdateRecords(t *testing.T) {
	client := setupTest(t, http.MethodPatch, "/v1/acme/zones/example.org/rrsets", http.StatusOK, "rrsets-response.json")

	rrSet := []UpdateRRSet{{
		Name:       "acme.example.org.",
		ChangeType: "add",
		Type:       "TXT",
		Records:    []Record{{Content: `"my-acme-challenge"`}},
	}}

	resp, err := client.UpdateRecords(context.Background(), "example.org", rrSet)
	require.NoError(t, err)

	expected := &APIResponse{Status: "ok", Message: "RRsets updated"}

	assert.Equal(t, expected, resp)
}
