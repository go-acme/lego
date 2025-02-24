package internal

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern string, status int, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if filename == "" {
			rw.WriteHeader(status)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient("token")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_UpdateDNSZone(t *testing.T) {
	client := setupTest(t, "PATCH /dnszone/example.com", http.StatusOK, "")

	updateRequest := DnszoneUpdateRequest{
		Add: []Content{
			{
				Name:    "@",
				Type:    "TXT",
				TTL:     60,
				Content: "value",
			},
		},
	}

	err := client.UpdateDNSZone(context.Background(), "example.com", updateRequest)
	require.NoError(t, err)
}

func TestClient_UpdateDNSZone_error(t *testing.T) {
	client := setupTest(t, "PATCH /dnszone/example.com", http.StatusUnprocessableEntity, "error.json")

	updateRequest := DnszoneUpdateRequest{
		Add: []Content{
			{
				Name:    "@",
				Type:    "TXT",
				TTL:     60,
				Content: "value",
			},
		},
	}

	err := client.UpdateDNSZone(context.Background(), "example.com", updateRequest)
	require.EqualError(t, err, "invalid_token: the supplied token is invalid")
}
