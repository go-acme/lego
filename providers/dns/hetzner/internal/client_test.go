package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, apiKey string) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(apiKey)
	client.baseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, mux
}

func TestClient_GetTxtRecord(t *testing.T) {
	const zoneID = "zoneA"
	const apiKey = "myKeyA"

	client, mux := setupTest(t, apiKey)

	mux.HandleFunc("/api/v1/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authHeader)
		if auth != apiKey {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
			return
		}

		zID := req.URL.Query().Get("zone_id")
		if zID != zoneID {
			http.Error(rw, fmt.Sprintf("invalid zone ID: %s", zID), http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/get_txt_record.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record, err := client.GetTxtRecord(t.Context(), "test1", "txttxttxt", zoneID)
	require.NoError(t, err)

	fmt.Println(record)
}

func TestClient_CreateRecord(t *testing.T) {
	const zoneID = "zoneA"
	const apiKey = "myKeyB"

	client, mux := setupTest(t, apiKey)

	mux.HandleFunc("/api/v1/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authHeader)
		if auth != apiKey {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
			return
		}

		file, err := os.Open("./fixtures/create_txt_record.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := DNSRecord{
		Name:   "test",
		Type:   "TXT",
		Value:  "txttxttxt",
		TTL:    600,
		ZoneID: zoneID,
	}

	err := client.CreateRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	const apiKey = "myKeyC"

	client, mux := setupTest(t, apiKey)

	mux.HandleFunc("/api/v1/records/recordID", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authHeader)
		if auth != apiKey {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
			return
		}
	})

	err := client.DeleteRecord(t.Context(), "recordID")
	require.NoError(t, err)
}

func TestClient_GetZoneID(t *testing.T) {
	const apiKey = "myKeyD"

	client, mux := setupTest(t, apiKey)

	mux.HandleFunc("/api/v1/zones", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authHeader)
		if auth != apiKey {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
			return
		}

		file, err := os.Open("./fixtures/get_zone_id.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	zoneID, err := client.GetZoneID(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, "zoneA", zoneID)
}
