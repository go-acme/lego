package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

	client, err := NewClient(server.URL, "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()

	mux.HandleFunc(pattern, handler)

	return client
}

func writeFixtureHandler(method, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		if req.Header.Get("X-Auth-Token") != "secret" {
			http.Error(rw, fmt.Sprintf("invalid token: %q", req.Header.Get("X-Auth-Token")), http.StatusUnauthorized)
			return
		}

		if filename == "" {
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, _ = io.Copy(rw, file)
	}
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "/domains/1234/records", writeFixtureHandler(http.MethodPost, "add-records.json"))

	err := client.AddRecord(t.Context(), "1234", Record{})
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/domains/1234/records", writeFixtureHandler(http.MethodDelete, ""))

	err := client.DeleteRecord(t.Context(), "1234", "2725233")
	require.NoError(t, err)
}

func TestClient_searchRecords(t *testing.T) {
	client := setupTest(t, "/domains/1234/records", writeFixtureHandler(http.MethodGet, "search-records.json"))

	records, err := client.searchRecords(t.Context(), "1234", "2725233", "A")
	require.NoError(t, err)

	expected := &Records{
		TotalEntries: 6,
		Records: []Record{
			{Name: "ftp.example.com", Type: "A", Data: "192.0.2.8", TTL: 5771, ID: "A-6817754"},
			{Name: "example.com", Type: "A", Data: "192.0.2.17", TTL: 86400, ID: "A-6822994"},
			{Name: "example.com", Type: "NS", Data: "ns.rackspace.com", TTL: 3600, ID: "NS-6251982"},
			{Name: "example.com", Type: "NS", Data: "ns2.rackspace.com", TTL: 3600, ID: "NS-6251983"},
			{Name: "example.com", Type: "MX", Data: "mail.example.com", TTL: 3600, ID: "MX-3151218"},
			{Name: "www.example.com", Type: "CNAME", Data: "example.com", TTL: 5400, ID: "CNAME-9778009"},
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_listDomainsByName(t *testing.T) {
	client := setupTest(t, "/domains", writeFixtureHandler(http.MethodGet, "list-domains-by-name.json"))

	domains, err := client.listDomainsByName(t.Context(), "1234")
	require.NoError(t, err)

	expected := &ZoneSearchResponse{
		TotalEntries: 114,
		HostedZones:  []HostedZone{{ID: "2725257", Name: "sub1.example.com"}},
	}

	assert.Equal(t, expected, domains)
}
