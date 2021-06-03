package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetRecords(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records", mockHandler(http.MethodGet, http.StatusOK, "get_records.json"))

	records, err := client.GetRecords("azone01")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:       1,
			Name:     "@",
			TTL:      3600,
			Data:     "ns1.simply.com",
			Type:     "NS",
			Priority: 0,
		},
		{
			ID:       2,
			Name:     "@",
			TTL:      3600,
			Data:     "ns2.simply.com",
			Type:     "NS",
			Priority: 0,
		},
		{
			ID:       3,
			Name:     "@",
			TTL:      3600,
			Data:     "ns3.simply.com",
			Type:     "NS",
			Priority: 0,
		},
		{
			ID:       4,
			Name:     "@",
			TTL:      3600,
			Data:     "ns4.simply.com",
			Type:     "NS",
			Priority: 0,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records", mockHandler(http.MethodGet, http.StatusBadRequest, "bad_auth_error.json"))

	records, err := client.GetRecords("azone01")
	require.Error(t, err)

	assert.Nil(t, records)
}

func TestClient_AddRecord(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records", mockHandler(http.MethodPost, http.StatusOK, "add_record.json"))

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	recordID, err := client.AddRecord("azone01", record)
	require.NoError(t, err)

	assert.EqualValues(t, 123456789, recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records", mockHandler(http.MethodPost, http.StatusNotFound, "bad_zone_error.json"))

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	recordID, err := client.AddRecord("azone01", record)
	require.Error(t, err)

	assert.Zero(t, recordID)
}

func TestClient_EditRecord(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records/123456789", mockHandler(http.MethodPut, http.StatusOK, "success.json"))

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	err := client.EditRecord("azone01", 123456789, record)
	require.NoError(t, err)
}

func TestClient_EditRecord_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records/123456789", mockHandler(http.MethodPut, http.StatusNotFound, "invalid_record_id.json"))

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	err := client.EditRecord("azone01", 123456789, record)
	require.Error(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records/123456789", mockHandler(http.MethodDelete, http.StatusOK, "success.json"))

	err := client.DeleteRecord("azone01", 123456789)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/accountname/apikey/my/products/azone01/dns/records/123456789", mockHandler(http.MethodDelete, http.StatusNotFound, "invalid_record_id.json"))

	err := client.DeleteRecord("azone01", 123456789)
	require.Error(t, err)
}

func setupTest(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("accountname", "apikey")
	require.NoError(t, err)

	client.baseURL, _ = url.Parse(server.URL)

	return mux, client
}

func mockHandler(method string, statusCode int, filename string) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		if filename == "" {
			rw.WriteHeader(statusCode)
			return
		}

		file, err := os.Open(filepath.FromSlash(path.Join("./fixtures", filename)))
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
