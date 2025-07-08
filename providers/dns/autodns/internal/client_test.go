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

		apiUser, apiKey, ok := req.BasicAuth()
		if apiUser != "user" || apiKey != "secret" || !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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

	client := NewClient("user", "secret", 123)
	client.HTTPClient = server.Client()
	client.BaseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddTxtRecords(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/zone/example.com/_stream", http.StatusOK, "add-record.json")

	records := []*ResourceRecord{{}}

	zone, err := client.AddTxtRecords(t.Context(), "example.com", records)
	require.NoError(t, err)

	expected := &Zone{
		Name: "example.com",
		ResourceRecords: []*ResourceRecord{{
			Name:  "example.com",
			TTL:   120,
			Type:  "TXT",
			Value: "txt",
			Pref:  1,
		}},
		Action:            "xxx",
		VirtualNameServer: "yyy",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_RemoveTXTRecords(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/zone/example.com/_stream", http.StatusOK, "add-record.json")

	records := []*ResourceRecord{{}}

	err := client.RemoveTXTRecords(t.Context(), "example.com", records)
	require.NoError(t, err)
}
