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

func setupTest(t *testing.T, pattern, method string, status int, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
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

	client, err := NewClient("secret", 123456)
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "/directory/v1/org/123456/domains/example.com/dns", http.MethodPost, http.StatusOK, "add-record.json")

	record := Record{
		Name: "_acme-challenge",
		Text: "txtxtxt",
		TTL:  60,
		Type: "TXT",
	}

	newRecord, err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:   789465,
		Name: "foo",
		Text: "_acme-challenge",
		TTL:  60,
		Type: "txtxtxt",
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "/directory/v1/org/123456/domains/example.com/dns", http.MethodGet, http.StatusUnauthorized, "error.json")

	record := Record{
		Name: "_acme-challenge",
		Text: "txtxtxt",
		TTL:  60,
		Type: "TXT",
	}

	newRecord, err := client.AddRecord(context.Background(), "example.com", record)
	require.Error(t, err)

	assert.Nil(t, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/directory/v1/org/123456/domains/example.com/dns/789456", http.MethodDelete, http.StatusOK, "delete-record.json")

	err := client.DeleteRecord(context.Background(), "example.com", 789456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "/directory/v1/org/123456/domains/example.com/dns/789456", http.MethodDelete, http.StatusUnauthorized, "error.json")

	err := client.DeleteRecord(context.Background(), "example.com", 789456)
	require.Error(t, err)
}
