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

func TestClient_GetZones(t *testing.T) {
	client := setupTest(t, "/user/v1/zones", testHandler(http.MethodGet, http.StatusOK, "zones.json"))

	zones, err := client.GetZones(t.Context(), "", 100, 0)
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:          10,
			Name:        "user1.ch",
			Email:       "test@hosttech.ch",
			TTL:         10800,
			Nameserver:  "ns1.hosttech.ch",
			Dnssec:      false,
			DnssecEmail: "test@hosttech.ch",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZones_error(t *testing.T) {
	client := setupTest(t, "/user/v1/zones", testHandler(http.MethodGet, http.StatusUnauthorized, "error.json"))

	_, err := client.GetZones(t.Context(), "", 100, 0)
	require.Error(t, err)
}

func TestClient_GetZone(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123", testHandler(http.MethodGet, http.StatusOK, "zone.json"))

	zone, err := client.GetZone(t.Context(), "123")
	require.NoError(t, err)

	expected := &Zone{
		ID:          10,
		Name:        "user1.ch",
		Email:       "test@hosttech.ch",
		TTL:         10800,
		Nameserver:  "ns1.hosttech.ch",
		Dnssec:      false,
		DnssecEmail: "test@hosttech.ch",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_GetZone_error(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123", testHandler(http.MethodGet, http.StatusUnauthorized, "error.json"))

	_, err := client.GetZone(t.Context(), "123")
	require.Error(t, err)
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123/records", testHandler(http.MethodGet, http.StatusOK, "records.json"))

	records, err := client.GetRecords(t.Context(), "123", "TXT")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      10,
			Type:    "A",
			Name:    "www",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      11,
			Type:    "AAAA",
			Name:    "www",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      12,
			Type:    "CAA",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      13,
			Type:    "CNAME",
			Name:    "www",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      14,
			Type:    "MX",
			Name:    "mail.example.com",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      14,
			Type:    "NS",
			Name:    "ns1.example.com",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      15,
			Type:    "PTR",
			Name:    "smtp.example.com",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      16,
			Type:    "SRV",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      17,
			Type:    "TXT",
			Text:    "v=spf1 ip4:1.2.3.4/32 -all",
			TTL:     3600,
			Comment: "my first record",
		},
		{
			ID:      17,
			Type:    "TLSA",
			Text:    "0 0 1 d2abde240d7cd3ee6b4b28c54df034b97983a1d16e8a410e4561cb106618e971",
			TTL:     3600,
			Comment: "my first record",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123/records", testHandler(http.MethodGet, http.StatusUnauthorized, "error.json"))

	_, err := client.GetRecords(t.Context(), "123", "TXT")
	require.Error(t, err)
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123/records", testHandler(http.MethodPost, http.StatusCreated, "record.json"))

	record := Record{
		Type:    "TXT",
		Name:    "lego",
		Text:    "content",
		TTL:     3600,
		Comment: "example",
	}

	newRecord, err := client.AddRecord(t.Context(), "123", record)
	require.NoError(t, err)

	expected := &Record{
		ID:      10,
		Type:    "TXT",
		Name:    "lego",
		Text:    "content",
		TTL:     3600,
		Comment: "example",
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123/records", testHandler(http.MethodPost, http.StatusUnauthorized, "error-details.json"))

	record := Record{
		Type:    "TXT",
		Name:    "lego",
		Text:    "content",
		TTL:     3600,
		Comment: "example",
	}

	_, err := client.AddRecord(t.Context(), "123", record)
	require.Error(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123/records/6", testHandler(http.MethodDelete, http.StatusUnauthorized, "error.json"))

	err := client.DeleteRecord(t.Context(), "123", "6")
	require.Error(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "/user/v1/zones/123/records/6", testHandler(http.MethodDelete, http.StatusNoContent, ""))

	err := client.DeleteRecord(t.Context(), "123", "6")
	require.NoError(t, err)
}

func setupTest(t *testing.T, path string, handler http.Handler) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.Handle(path, handler)

	client := NewClient(OAuthStaticAccessToken(server.Client(), testAPIKey))
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func testHandler(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf(`{"message":"unsupported method: %s"}`, req.Method), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Authorization") != "Bearer "+testAPIKey {
			http.Error(rw, `{"message":"Unauthenticated"}`, http.StatusUnauthorized)
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
