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

const apiKey = "secret"

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(apiKey)
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func testHandler(filename, method string, statusCode int) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		username, key, ok := req.BasicAuth()
		if username != "api" || key != apiKey || !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
	client, mux := setupTest(t)

	mux.HandleFunc("/domains.json", testHandler("get-domains.json", http.MethodGet, http.StatusOK))

	domains, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{{
		ID:          123,
		UnicodeFqdn: "example.com",
		Domain:      "example",
		TLD:         "com",
		Status:      "ok",
	}}
	assert.Equal(t, expected, domains)
}

func TestClient_GetDomains_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains.json", testHandler("error.json", http.MethodGet, http.StatusBadRequest))

	_, err := client.GetDomains(t.Context())
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_GetRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records.json", testHandler("get-records.json", http.MethodGet, http.StatusOK))

	records, err := client.GetRecords(t.Context(), 123)
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      1234,
			Content: "ns1.lima-city.de",
			Name:    "example.com",
			TTL:     36000,
			Type:    "NS",
		},
		{
			ID:      5678,
			Content: `"foobar"`,
			Name:    "_acme-challenge.example.com",
			TTL:     36000,
			Type:    "TXT",
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records.json", testHandler("error.json", http.MethodGet, http.StatusBadRequest))

	_, err := client.GetRecords(t.Context(), 123)
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records.json", testHandler("ok.json", http.MethodPost, http.StatusOK))

	record := Record{
		Name:    "foo",
		Content: "bar",
		TTL:     12,
		Type:    "TXT",
	}

	err := client.AddRecord(t.Context(), 123, record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records.json", testHandler("error.json", http.MethodPost, http.StatusBadRequest))

	record := Record{
		Name:    "foo",
		Content: "bar",
		TTL:     12,
		Type:    "TXT",
	}

	err := client.AddRecord(t.Context(), 123, record)
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_UpdateRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records/456", testHandler("ok.json", http.MethodPut, http.StatusOK))

	err := client.UpdateRecord(t.Context(), 123, 456, Record{})
	require.NoError(t, err)
}

func TestClient_UpdateRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records/456", testHandler("error.json", http.MethodPut, http.StatusBadRequest))

	err := client.UpdateRecord(t.Context(), 123, 456, Record{})
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records/456", testHandler("ok.json", http.MethodDelete, http.StatusOK))

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/123/records/456", testHandler("error.json", http.MethodDelete, http.StatusBadRequest))

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.EqualError(t, err, "[status code: 400] status: invalid_resource, details: name: [muss ausgefüllt werden]")
}
