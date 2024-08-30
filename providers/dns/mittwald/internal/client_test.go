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

func setupTest(t *testing.T, pattern string, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, handler)

	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func testHandler(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf(`{"message":"unsupported method: %s"}`, req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authorizationHeader)
		if auth != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid API Token: %s", auth), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(statusCode)

		if statusCode == http.StatusNoContent {
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, fmt.Sprintf(`{"message":"%v"}`, err), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, fmt.Sprintf(`{"message":"%v"}`, err), http.StatusInternalServerError)
			return
		}
	}
}

func TestClient_ListDomains(t *testing.T) {
	client := setupTest(t, "/domains", testHandler(http.MethodGet, http.StatusOK, "domain-list-domains.json"))

	domains, err := client.ListDomains(context.Background())
	require.NoError(t, err)

	require.Len(t, domains, 1)

	expected := []Domain{{
		Domain:    "string",
		DomainID:  "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		ProjectID: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
	}}

	assert.Equal(t, expected, domains)
}

func TestClient_ListDomains_error(t *testing.T) {
	client := setupTest(t, "/domains", testHandler(http.MethodGet, http.StatusBadRequest, "error-client.json"))

	_, err := client.ListDomains(context.Background())
	require.EqualError(t, err, "[status code 400] ValidationError: Validation failed [format: should be string (.address.street, email)]")
}

func TestClient_ListDNSZones(t *testing.T) {
	client := setupTest(t, "/projects/my-project-id/dns-zones", testHandler(http.MethodGet, http.StatusOK, "dns-list-dns-zones.json"))

	zones, err := client.ListDNSZones(context.Background(), "my-project-id")
	require.NoError(t, err)

	require.Len(t, zones, 1)

	expected := []DNSZone{{
		ID:     "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		Domain: "string",
		RecordSet: &RecordSet{
			TXT: &TXTRecord{},
		},
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_GetDNSZone(t *testing.T) {
	client := setupTest(t, "/dns-zones/my-zone-id", testHandler(http.MethodGet, http.StatusOK, "dns-get-dns-zone.json"))

	zone, err := client.GetDNSZone(context.Background(), "my-zone-id")
	require.NoError(t, err)

	expected := &DNSZone{
		ID:     "3fa85f64-5717-4562-b3fc-2c963f66afa6",
		Domain: "string",
		RecordSet: &RecordSet{
			TXT: &TXTRecord{},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_CreateDNSZone(t *testing.T) {
	client := setupTest(t, "/dns-zones", testHandler(http.MethodPost, http.StatusCreated, "dns-create-dns-zone.json"))

	request := CreateDNSZoneRequest{
		Name:         "test",
		ParentZoneID: "my-parent-zone-id",
	}

	zone, err := client.CreateDNSZone(context.Background(), request)
	require.NoError(t, err)

	expected := &DNSZone{
		ID: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_UpdateTXTRecord(t *testing.T) {
	client := setupTest(t, "/dns-zones/my-zone-id/record-sets/txt", testHandler(http.MethodPut, http.StatusNoContent, ""))

	record := TXTRecord{
		Settings: Settings{
			TTL: TTL{Auto: true},
		},
		Entries: []string{"txt"},
	}

	err := client.UpdateTXTRecord(context.Background(), "my-zone-id", record)
	require.NoError(t, err)
}

func TestClient_DeleteDNSZone(t *testing.T) {
	client := setupTest(t, "/dns-zones/my-zone-id", testHandler(http.MethodDelete, http.StatusOK, ""))

	err := client.DeleteDNSZone(context.Background(), "my-zone-id")
	require.NoError(t, err)
}

func TestClient_DeleteDNSZone_error(t *testing.T) {
	client := setupTest(t, "/dns-zones/my-zone-id", testHandler(http.MethodDelete, http.StatusInternalServerError, "error.json"))

	err := client.DeleteDNSZone(context.Background(), "my-zone-id")
	assert.EqualError(t, err, "[status code 500] InternalServerError: Something went wrong")
}
