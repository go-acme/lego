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

	client := NewClient("key", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return mux, client
}

func TestClient_GetRecords(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/v1/domains/example.com/records/TXT/", testHandler(http.MethodGet, http.StatusOK, "getrecords.json"))

	records, err := client.GetRecords("example.com", "TXT", "")
	require.NoError(t, err)

	expected := []DNSRecord{
		{Name: "_acme-challenge", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "6rrai7-jm7l3PxL4WGmGoS6VMeefSHx24r-qCvUIOxU", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "8Axt-PXQvjOVD2oi2YXqyyn8U5CDcC8P-BphlCxk3Ek", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "0Ad60wO_yxxJPFPb1deir6lQ37FPLeA02YCobo7Qm8A", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "acme", TTL: 600},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_errors(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/v1/domains/example.com/records/TXT/", testHandler(http.MethodGet, http.StatusUnprocessableEntity, "errors.json"))

	records, err := client.GetRecords("example.com", "TXT", "")
	require.Error(t, err)
	assert.Nil(t, records)
}

func TestClient_UpdateTxtRecords(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/v1/domains/example.com/records/TXT/lego", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut {
			http.Error(rw, fmt.Sprintf(`{"message":"unsupported method: %s"}`, req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "sso-key key:secret" {
			http.Error(rw, fmt.Sprintf("invalid API key or secret: %s", auth), http.StatusUnauthorized)
			return
		}
	})

	records := []DNSRecord{
		{Name: "_acme-challenge", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "6rrai7-jm7l3PxL4WGmGoS6VMeefSHx24r-qCvUIOxU", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "8Axt-PXQvjOVD2oi2YXqyyn8U5CDcC8P-BphlCxk3Ek", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "0Ad60wO_yxxJPFPb1deir6lQ37FPLeA02YCobo7Qm8A", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "acme", TTL: 600},
	}

	err := client.UpdateTxtRecords(records, "example.com", "lego")
	require.NoError(t, err)
}

func TestClient_UpdateTxtRecords_errors(t *testing.T) {
	mux, client := setupTest(t)

	mux.HandleFunc("/v1/domains/example.com/records/TXT/lego",
		testHandler(http.MethodPut, http.StatusUnprocessableEntity, "errors.json"))

	records := []DNSRecord{
		{Name: "_acme-challenge", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "6rrai7-jm7l3PxL4WGmGoS6VMeefSHx24r-qCvUIOxU", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "8Axt-PXQvjOVD2oi2YXqyyn8U5CDcC8P-BphlCxk3Ek", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "0Ad60wO_yxxJPFPb1deir6lQ37FPLeA02YCobo7Qm8A", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "acme", TTL: 600},
	}

	err := client.UpdateTxtRecords(records, "example.com", "lego")
	require.Error(t, err)
}

func testHandler(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf(`{"message":"unsupported method: %s"}`, req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "sso-key key:secret" {
			http.Error(rw, fmt.Sprintf("invalid API key or secret: %s", auth), http.StatusUnauthorized)
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
