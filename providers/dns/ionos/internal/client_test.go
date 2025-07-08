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

func TestClient_ListZones(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones", mockHandler(http.MethodGet, http.StatusOK, "list_zones.json"))

	zones, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{{
		ID:   "11af3414-ebba-11e9-8df5-66fbe8a334b4",
		Name: "test.com",
		Type: "NATIVE",
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones", mockHandler(http.MethodGet, http.StatusUnauthorized, "list_zones_error.json"))

	zones, err := client.ListZones(t.Context())
	require.Error(t, err)

	assert.Nil(t, zones)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusUnauthorized, cErr.StatusCode)
}

func TestClient_GetRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones/azone01", mockHandler(http.MethodGet, http.StatusOK, "get_records.json"))

	records, err := client.GetRecords(t.Context(), "azone01", nil)
	require.NoError(t, err)

	expected := []Record{{
		ID:      "22af3414-abbe-9e11-5df5-66fbe8e334b4",
		Name:    "string",
		Content: "string",
		Type:    "A",
	}}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones/azone01", mockHandler(http.MethodGet, http.StatusUnauthorized, "get_records_error.json"))

	records, err := client.GetRecords(t.Context(), "azone01", nil)
	require.Error(t, err)

	assert.Nil(t, records)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusUnauthorized, cErr.StatusCode)
}

func TestClient_RemoveRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones/azone01/records/arecord01", mockHandler(http.MethodDelete, http.StatusOK, ""))

	err := client.RemoveRecord(t.Context(), "azone01", "arecord01")
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones/azone01/records/arecord01", mockHandler(http.MethodDelete, http.StatusInternalServerError, "remove_record_error.json"))

	err := client.RemoveRecord(t.Context(), "azone01", "arecord01")
	require.Error(t, err)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusInternalServerError, cErr.StatusCode)
}

func TestClient_ReplaceRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones/azone01", mockHandler(http.MethodPatch, http.StatusOK, ""))

	records := []Record{{
		ID:      "22af3414-abbe-9e11-5df5-66fbe8e334b4",
		Name:    "string",
		Content: "string",
		Type:    "A",
	}}

	err := client.ReplaceRecords(t.Context(), "azone01", records)
	require.NoError(t, err)
}

func TestClient_ReplaceRecords_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/zones/azone01", mockHandler(http.MethodPatch, http.StatusBadRequest, "replace_records_error.json"))

	records := []Record{{
		ID:      "22af3414-abbe-9e11-5df5-66fbe8e334b4",
		Name:    "string",
		Content: "string",
		Type:    "A",
	}}

	err := client.ReplaceRecords(t.Context(), "azone01", records)
	require.Error(t, err)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusBadRequest, cErr.StatusCode)
}

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("secret")
	require.NoError(t, err)

	client.BaseURL, _ = url.Parse(server.URL)

	return client, mux
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
