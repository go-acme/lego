package internal

import (
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
	client := setupTest(t, "PATCH /dnszone/example.com", http.StatusOK, "update-dns-zone.json")

	updateRequest := DNSZoneUpdateRequest{
		Add: []Record{{
			Name:    "@",
			Type:    "TXT",
			TTL:     60,
			Content: "value",
		}},
	}

	response, err := client.UpdateDNSZone(t.Context(), "example.com", updateRequest)
	require.NoError(t, err)

	expected := &DNSZoneUpdateResponse{
		ResponseID: "mapi1_cb46ad8790b62b76535bd3102bd282aec83b894c",
		Status:     "ok",
		Message:    "Command completed successfully",
	}

	assert.Equal(t, expected, response)
}

func TestClient_UpdateDNSZone_error(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		expected string
	}{
		{
			desc:     "authentication error",
			filename: "error.json",
			expected: "invalid_token: the supplied token is invalid",
		},
		{
			desc:     "API error",
			filename: "error-response.json",
			expected: "error: does_not_exist: This server does not exist",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := setupTest(t, "PATCH /dnszone/example.com", http.StatusUnprocessableEntity, test.filename)

			updateRequest := DNSZoneUpdateRequest{
				Add: []Record{{
					Name:    "@",
					Type:    "TXT",
					TTL:     60,
					Content: "value",
				}},
			}

			_, err := client.UpdateDNSZone(t.Context(), "example.com", updateRequest)
			require.EqualError(t, err, test.expected)
		})
	}
}
