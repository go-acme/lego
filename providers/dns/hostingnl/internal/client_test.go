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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, method string, status int, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/domains/example.com/dns", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		open, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, http.MethodPost, http.StatusOK, "add-record.json")

	record := Record{
		Name:     "example.com",
		Type:     "TXT",
		Content:  strconv.Quote("txtxtxt"),
		TTL:      "3600",
		Priority: "0",
	}

	newRecord, err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:       "12345",
		Name:     "example.com",
		Type:     "TXT",
		Content:  `"txtxtxt"`,
		TTL:      "3600",
		Priority: "0",
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, http.MethodDelete, http.StatusOK, "delete-record.json")

	err := client.DeleteRecord(context.Background(), "example.com", "12345")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, http.MethodDelete, http.StatusUnauthorized, "error.json")

	err := client.DeleteRecord(context.Background(), "example.com", "12345")
	require.Error(t, err)
}

func TestClient_DeleteRecord_error_other(t *testing.T) {
	client := setupTest(t, http.MethodDelete, http.StatusNotFound, "error-other.json")

	err := client.DeleteRecord(context.Background(), "example.com", "12345")
	require.Error(t, err)
}
