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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetRecords(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/dns/example.com/records", testHandler("./GetRecords.json", http.MethodGet, http.StatusOK))

	records, err := client.GetRecords(context.Background(), "example.com", nil)
	require.NoError(t, err)

	expected := []Record{
		{
			ID:         "string",
			Type:       "string",
			RecordName: "string",
			Content:    "string",
			TTL:        3600,
			Priority:   10,
			Service:    "string",
			Weight:     0,
			Target:     "string",
			Protocol:   "TCP",
			Port:       0,
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_GetRecord(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/dns/example.com/records/123", testHandler("./GetRecord.json", http.MethodGet, http.StatusOK))

	record, err := client.GetRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)

	expected := &Record{
		ID:         "string",
		Type:       "string",
		RecordName: "string",
		Content:    "string",
		TTL:        3600,
		Priority:   10,
		Service:    "string",
		Weight:     0,
		Target:     "string",
		Protocol:   "TCP",
		Port:       0,
	}
	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/dns/example.com/records/123", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/dns/example.com/records/123", testHandler("./error.json", http.MethodDelete, http.StatusUnauthorized))

	err := client.DeleteRecord(context.Background(), "example.com", "123")
	require.Error(t, err)
}

func TestClient_sign(t *testing.T) {
	client := NewClient("my_key", "my_secret", nil)

	endpoint, err := url.Parse("https://localhost/v2/domains")
	require.NoError(t, err)

	query := endpoint.Query()
	query.Set("skip", "0")
	query.Set("take", "10")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	require.NoError(t, err)

	sign, err := client.sign(req, nil)
	require.NoError(t, err)

	assert.Regexp(t, `hmac my_key:[^:]+:[a-zA-Z]{10}:\d{10}`, sign)
}

func testHandler(filename string, method string, statusCode int) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "hmac key:") {
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

func setup(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("key", "secret", nil)
	client.httpClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}
