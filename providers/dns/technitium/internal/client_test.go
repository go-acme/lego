package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		file, err := os.Open(filepath.Join("fixtures", filename))
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

	client, err := NewClient(server.URL, "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()

	return client
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "POST /api/zones/records/add", "add-record.json")

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	newRecord, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	expected := &Record{Name: "example.com", Type: "A"}

	assert.Equal(t, expected, newRecord)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "POST /api/zones/records/add", "error.json")

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	_, err := client.AddRecord(t.Context(), record)
	require.Error(t, err)

	assert.EqualError(t, err, "Status: error, ErrorMessage: error message, StackTrace: application stack trace, InnerErrorMessage: inner exception message")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "POST /api/zones/records/delete", "delete-record.json")

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	err := client.DeleteRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "POST /api/zones/records/delete", "error.json")

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	err := client.DeleteRecord(t.Context(), record)
	require.Error(t, err)

	assert.EqualError(t, err, "Status: error, ErrorMessage: error message, StackTrace: application stack trace, InnerErrorMessage: inner exception message")
}
