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

func setupTest(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return mux, client
}

func TestClient_GetDNSRecords(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/domains/example.com/records", testHandler(http.MethodGet, http.StatusOK, "getDnsRecord.json"))

	records, err := client.GetDNSRecords("example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:   "abc123",
			Name: "www",
			Type: "CAA",
			Data: "1 issue letsencrypt.org",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Name: "www",
			Type: "A",
			Data: "192.64.147.249",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Name: "*",
			Type: "A",
			Data: "192.64.147.249",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Type: "CAA",
			Data: "0 issue trust-provider.com",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Type: "CAA",
			Data: "1 issue letsencrypt.org",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Type: "A",
			Data: "192.64.147.249",
			AUX:  0,
			TTL:  300,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetDNSRecords_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/domains/example.com/records", testHandler(http.MethodGet, http.StatusUnauthorized, "error.json"))

	_, err := client.GetDNSRecords("example.com")
	assert.Error(t, err)
}

func TestClient_CreateHostRecord(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/domains/example.com/records", testHandler(http.MethodPost, http.StatusOK, "createHostRecord.json"))

	record := RecordRequest{
		Host: "www2",
		Type: "A",
		Data: "192.64.147.249",
		Aux:  0,
		TTL:  300,
	}

	data, err := client.CreateHostRecord("example.com", record)
	require.NoError(t, err)

	expected := &Data{
		Code:    1000,
		Message: "Command completed successfully.",
	}

	assert.Equal(t, expected, data)
}

func TestClient_CreateHostRecord_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/domains/example.com/records", testHandler(http.MethodPost, http.StatusUnauthorized, "error.json"))

	record := RecordRequest{
		Host: "www2",
		Type: "A",
		Data: "192.64.147.249",
		Aux:  0,
		TTL:  300,
	}

	_, err := client.CreateHostRecord("example.com", record)
	assert.Error(t, err)
}

func TestClient_RemoveHostRecord(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/domains/example.com/records", testHandler(http.MethodDelete, http.StatusOK, "removeHostRecord.json"))

	data, err := client.RemoveHostRecord("example.com", "abc123")
	require.NoError(t, err)

	expected := &Data{
		Code:    1000,
		Message: "Command completed successfully.",
	}

	assert.Equal(t, expected, data)
}

func TestClient_RemoveHostRecord_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/domains/example.com/records", testHandler(http.MethodDelete, http.StatusUnauthorized, "error.json"))

	_, err := client.RemoveHostRecord("example.com", "abc123")
	assert.Error(t, err)
}

func testHandler(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf(`{"message":"unsupported method: %s"}`, req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.URL.Query().Get("SIGNATURE")
		if auth != "secret" {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
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
