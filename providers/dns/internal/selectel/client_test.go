package selectel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListRecords(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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

	client := NewClient("token")
	client.BaseURL = server.URL

	records, err := client.ListRecords(123)
	require.NoError(t, err)

	expected := []Record{
		{ID: 123, Name: "example.com", Type: "TXT", TTL: 60, Email: "email@example.com", Content: "txttxttxtA"},
		{ID: 1234, Name: "example.org", Type: "TXT", TTL: 60, Email: "email@example.org", Content: "txttxttxtB"},
		{ID: 12345, Name: "example.net", Type: "TXT", TTL: 60, Email: "email@example.net", Content: "txttxttxtC"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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

	client := NewClient("token")
	client.BaseURL = server.URL

	records, err := client.ListRecords(123)

	assert.EqualError(t, err, "request failed with status code 401: API error: 400 - error description - field that the error occurred in")
	assert.Nil(t, records)
}

func TestClient_GetDomainByName(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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

	client := NewClient("token")
	client.BaseURL = server.URL

	domain, err := client.GetDomainByName("sub.sub.example.org")
	require.NoError(t, err)

	expected := &Domain{
		ID:   123,
		Name: "example.org",
	}

	assert.Equal(t, expected, domain)
}

func TestClient_AddRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

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

	client := NewClient("token")
	client.BaseURL = server.URL

	record, err := client.AddRecord(123, Record{
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
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}
	})

	client := NewClient("token")
	client.BaseURL = server.URL

	err := client.DeleteRecord(123, 456)
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
