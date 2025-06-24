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

const testAPIKey = "secret"

func setupTest(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)

	client := NewClient(OAuthStaticAccessToken(server.Client(), testAPIKey))
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func testHandler(method, filename string, statusCode int) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer "+testAPIKey {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(statusCode)

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func TestClient_GetDomains(t *testing.T) {
	client := setupTest(t, testHandler(http.MethodGet, "get_domains.json", http.StatusOK))

	domains, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := &Domains{
		APIResponse: APIResponse{
			Status: "200 OK",
			Info:   "success",
		},
		APICall: "get_domains",
		Subdomains: map[string]Subdomain{
			"lego.home64.net": {
				Updates:          0,
				Wildcard:         1,
				DomainUpdateHash: "Dr4l6jFVgkXITqZPEyMHLNsGAfwoSu9v",
				Records: []Record{
					{
						RecordID:   50665,
						Content:    "2606:2800:220:1:248:1893:25c8:1946",
						TTL:        60,
						Type:       "AAAA",
						Prefix:     "",
						LastUpdate: "2023-07-19 13:18:59",
						RecordKey:  "MTA0YzdmMWVjYTFiNDBmZjYwMTU0OGUy",
					},
				},
			},
			"lego.ipv64.net": {
				Updates:          0,
				Wildcard:         1,
				DomainUpdateHash: "Dr4l6jFVgkXITqZPEyMHLNsGAfwoSu9v",
				Records: []Record{
					{
						RecordID:   50664,
						Content:    "2606:2800:220:1:248:1893:25c8:1946",
						TTL:        60,
						Type:       "AAAA",
						Prefix:     "",
						LastUpdate: "2023-07-19 13:18:59",
						RecordKey:  "ZDMxOWUxMjZjOTk5MmQ3N2M3ODc4NjJj",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_GetDomains_error(t *testing.T) {
	client := setupTest(t, testHandler(http.MethodGet, "error.json", http.StatusUnauthorized))

	domains, err := client.GetDomains(t.Context())
	require.Error(t, err)

	require.Nil(t, domains)
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, testHandler(http.MethodPost, "add_record.json", http.StatusCreated))

	err := client.AddRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, testHandler(http.MethodPost, "add_record-error.json", http.StatusBadRequest))

	err := client.AddRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.Error(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, testHandler(http.MethodDelete, "del_record.json", http.StatusAccepted))

	err := client.DeleteRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, testHandler(http.MethodDelete, "del_record-error.json", http.StatusBadRequest))

	err := client.DeleteRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.Error(t, err)
}
