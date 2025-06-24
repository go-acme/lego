package selectel

import (
	"encoding/json"
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

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("token")
	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, mux
}

func TestClient_ListRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/123/records/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		fixture := "./fixtures/list_records.json"

		err := writeResponse(rw, fixture)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := client.ListRecords(t.Context(), 123)
	require.NoError(t, err)

	expected := []Record{
		{ID: 123, Name: "example.com", Type: "TXT", TTL: 60, Email: "email@example.com", Content: "txttxttxtA"},
		{ID: 1234, Name: "example.org", Type: "TXT", TTL: 60, Email: "email@example.org", Content: "txttxttxtB"},
		{ID: 12345, Name: "example.net", Type: "TXT", TTL: 60, Email: "email@example.net", Content: "txttxttxtC"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/123/records/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		rw.WriteHeader(http.StatusUnauthorized)
		err := writeResponse(rw, "./fixtures/error.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := client.ListRecords(t.Context(), 123)

	require.EqualError(t, err, "request failed with status code 401: API error: 400 - error description - field that the error occurred in")
	assert.Nil(t, records)
}

func TestClient_GetDomainByName(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/sub.sub.example.org", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		rw.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/sub.example.org", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		rw.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/example.org", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		fixture := "./fixtures/domains.json"

		err := writeResponse(rw, fixture)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	domain, err := client.GetDomainByName(t.Context(), "sub.sub.example.org")
	require.NoError(t, err)

	expected := &Domain{
		ID:   123,
		Name: "example.org",
	}

	assert.Equal(t, expected, domain)
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/123/records/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		rec := Record{}

		err := json.NewDecoder(req.Body).Decode(&rec)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rec.ID = 456

		err = json.NewEncoder(rw).Encode(rec)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record, err := client.AddRecord(t.Context(), 123, Record{
		Name:    "example.org",
		Type:    "TXT",
		TTL:     60,
		Email:   "email@example.org",
		Content: "txttxttxttxt",
	})

	require.NoError(t, err)

	expected := &Record{
		ID:      456,
		Name:    "example.org",
		Type:    "TXT",
		TTL:     60,
		Email:   "email@example.org",
		Content: "txttxttxttxt",
	}

	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}
	})

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.NoError(t, err)
}

func writeResponse(rw io.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	_, err = io.Copy(rw, file)
	return err
}
