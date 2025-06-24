package whm

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		open, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(http.StatusOK)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient(server.URL, "user", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()

	return client
}

func TestClient_FetchZoneInformation(t *testing.T) {
	client := setupTest(t, "/json-api/parse_dns_zone", "zone-info.json")

	zoneInfo, err := client.FetchZoneInformation(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []shared.ZoneRecord{{
		LineIndex:  22,
		Type:       "record",
		DataB64:    []string{"dGV4YXMuY29tLg=="},
		DNameB64:   "dGV4YXMuY29tLg==",
		RecordType: "MX",
		TTL:        14400,
	}}

	assert.Equal(t, expected, zoneInfo)
}

func TestClient_FetchZoneInformation_error(t *testing.T) {
	client := setupTest(t, "/json-api/parse_dns_zone", "zone-info_error.json")

	zoneInfo, err := client.FetchZoneInformation(t.Context(), "example.com")
	require.Error(t, err)

	assert.Nil(t, zoneInfo)
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "/json-api/mass_edit_dns_zone", "update-zone.json")

	record := shared.Record{
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.AddRecord(t.Context(), 123456, "example.com", record)
	require.NoError(t, err)

	expected := &shared.ZoneSerial{NewSerial: "2021031903"}

	assert.Equal(t, expected, zoneSerial)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "/json-api/mass_edit_dns_zone", "update-zone_error.json")

	record := shared.Record{
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.AddRecord(t.Context(), 123456, "example.com", record)
	require.Error(t, err)

	assert.Nil(t, zoneSerial)
}

func TestClient_EditRecord(t *testing.T) {
	client := setupTest(t, "/json-api/mass_edit_dns_zone", "update-zone.json")

	record := shared.Record{
		LineIndex:  9,
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.EditRecord(t.Context(), 123456, "example.com", record)
	require.NoError(t, err)

	expected := &shared.ZoneSerial{NewSerial: "2021031903"}

	assert.Equal(t, expected, zoneSerial)
}

func TestClient_EditRecord_error(t *testing.T) {
	client := setupTest(t, "/json-api/mass_edit_dns_zone", "update-zone_error.json")

	record := shared.Record{
		LineIndex:  9,
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.EditRecord(t.Context(), 123456, "example.com", record)
	require.Error(t, err)

	assert.Nil(t, zoneSerial)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/json-api/mass_edit_dns_zone", "update-zone.json")

	zoneSerial, err := client.DeleteRecord(t.Context(), 123456, "example.com", 0)
	require.NoError(t, err)

	expected := &shared.ZoneSerial{NewSerial: "2021031903"}

	assert.Equal(t, expected, zoneSerial)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "/json-api/mass_edit_dns_zone", "update-zone_error.json")

	zoneSerial, err := client.DeleteRecord(t.Context(), 123456, "example.com", 0)
	require.Error(t, err)

	assert.Nil(t, zoneSerial)
}
