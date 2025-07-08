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

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("user", "secret")
	client.baseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, mux
}

func testHandler(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf(`unsupported method: %s`, req.Method), http.StatusMethodNotAllowed)
			return
		}

		rw.WriteHeader(statusCode)

		if statusCode == http.StatusNoContent {
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, fmt.Sprintf(`message %v`, err), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, fmt.Sprintf(`message %v`, err), http.StatusInternalServerError)
			return
		}
	}
}

func testHandlerAuth(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf(`{"message":"unsupported method: %s"}`, req.Method), http.StatusMethodNotAllowed)
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

func TestClient_CreateAuthenticationToken(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/auth/token", testHandlerAuth(http.MethodPost, http.StatusOK, "auth.json"))

	ctx := t.Context()

	token, err := client.CreateAuthenticationToken(ctx)
	require.NoError(t, err)

	expected := &Token{
		Token:   "authsecret",
		Expires: 123,
	}
	assert.Equal(t, expected, token)
}

func TestClient_ListZone(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dnszones/", testHandler(http.MethodGet, http.StatusOK, "ListZone.json"))

	ctx := t.Context()

	zones, err := client.ListZone(ctx)
	require.NoError(t, err)

	expected := []Zone{
		{Name: "example.com", Type: "master"},
		{Name: "example.net", Type: "slave"},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZoneDetails(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dnszones/example.com", testHandler(http.MethodGet, http.StatusOK, "GetZoneDetails.json"))

	ctx := t.Context()

	zone, err := client.GetZoneDetails(ctx, "example.com")
	require.NoError(t, err)

	expected := &ZoneDetails{
		Active: true,
		DNSSec: true,
		Name:   "example.com",
		Type:   "master",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_ListRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dnszones/example.com/records/", testHandler(http.MethodGet, http.StatusOK, "ListRecords.json"))

	ctx := t.Context()

	records, err := client.ListRecords(ctx, "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			Name: "@",
			TTL:  86400,
			Type: "NS",
			Data: "ns2.core-networks.eu.",
		},
		{
			Name: "@",
			TTL:  86400,
			Type: "NS",
			Data: "ns3.core-networks.com.",
		},
		{
			Name: "@",
			TTL:  86400,
			Type: "NS",
			Data: "ns1.core-networks.de.",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dnszones/example.com/records/", testHandler(http.MethodPost, http.StatusNoContent, ""))

	ctx := t.Context()

	record := Record{Name: "www", TTL: 3600, Type: "A", Data: "127.0.0.1"}

	err := client.AddRecord(ctx, "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dnszones/example.com/records/delete", testHandler(http.MethodPost, http.StatusNoContent, ""))

	ctx := t.Context()

	record := Record{Name: "www", Type: "A", Data: "127.0.0.1"}

	err := client.DeleteRecords(ctx, "example.com", record)
	require.NoError(t, err)
}

func TestClient_CommitRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dnszones/example.com/records/commit", testHandler(http.MethodPost, http.StatusNoContent, ""))

	ctx := t.Context()

	err := client.CommitRecords(ctx, "example.com")
	require.NoError(t, err)
}
