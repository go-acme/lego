package internal

import (
	"context"
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

const apiKey = "key"

func TestClient_GetRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/v1/zones/example.com/dns-records", testHandler("./RecordsResponse.json", http.MethodGet, http.StatusOK))

	records, err := client.GetRecords(context.Background(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:   "string",
			Type: "string",
			Name: "string",
			Contents: []Content{
				{
					Text: "string",
				},
			},
			TTL: 3600,
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_GetRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/v1/zones/example.com/dns-records/123", testHandler("./RecordResponse.json", http.MethodGet, http.StatusOK))

	record, err := client.GetRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)

	expected := &Record{
		ID:   "string",
		Type: "string",
		Name: "string",
		Contents: []Content{
			{
				Text: "string",
			},
		},
		TTL: 3600,
	}
	assert.Equal(t, expected, record)
}

func TestClient_CreateRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/v1/zones/example.com/dns-records", testHandler("./RecordResponse.json", http.MethodPost, http.StatusCreated))

	data := Record{
		Type: "string",
		Name: "string",
		Contents: []Content{
			{
				Text: "string",
			},
		},
		TTL: 3600,
	}

	record, err := client.CreateRecord(context.Background(), "example.com", data)
	require.NoError(t, err)

	expected := &Record{
		ID:   "string",
		Type: "string",
		Name: "string",
		Contents: []Content{
			{
				Text: "string",
			},
		},
		TTL: 3600,
	}

	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/v1/zones/example.com/dns-records/123", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_NotFound_Response(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/v1/zones/example.com/dns-records/123", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	})

	err := client.DeleteRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/api/v1/zones/example.com/dns-records/123", testHandler("./error.json", http.MethodDelete, http.StatusUnauthorized))

	err := client.DeleteRecord(context.Background(), "example.com", "123")
	require.Error(t, err)
}

func testHandler(filename string, method string, statusCode int) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer "+apiKey {
			http.Error(rw, "invalid Authorization header", http.StatusUnauthorized)
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

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(OAuthStaticAccessToken(server.Client(), apiKey))
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}
