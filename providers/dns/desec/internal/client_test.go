package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTxtRRSet(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("token")
	client.BaseURL = server.URL

	mux.HandleFunc("/domains/example.dedyn.io/rrsets/_acme-challenge/TXT/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open("./fixtures/get_record.json")
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

	record, err := client.GetTxtRRSet("example.dedyn.io", "_acme-challenge")
	require.NoError(t, err)

	expected := &RRSet{
		Name:    "_acme-challenge.example.dedyn.io.",
		Domain:  "example.dedyn.io",
		SubName: "_acme-challenge",
		Type:    "TXT",
		Records: []string{`"txt"`},
		TTL:     300,
	}
	assert.Equal(t, expected, record)
}

func TestAddTxtRRSet(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("token")
	client.BaseURL = server.URL

	mux.HandleFunc("/domains/example.dedyn.io/rrsets/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		rw.WriteHeader(http.StatusCreated)
		file, err := os.Open("./fixtures/add_record.json")
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

	record := RRSet{
		Name:    "",
		Domain:  "example.dedyn.io",
		SubName: "_acme-challenge",
		Type:    "TXT",
		Records: []string{`"txt"`},
		TTL:     300,
	}

	newRecord, err := client.AddTxtRRSet(record)
	require.NoError(t, err)

	expected := &RRSet{
		Name:    "_acme-challenge.example.dedyn.io.",
		Domain:  "example.dedyn.io",
		SubName: "_acme-challenge",
		Type:    "TXT",
		Records: []string{`"txt"`},
		TTL:     300,
	}
	assert.Equal(t, expected, newRecord)
}

func TestUpdateTxtRRSet(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("token")
	client.BaseURL = server.URL

	mux.HandleFunc("/domains/example.dedyn.io/rrsets/_acme-challenge/TXT/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open("./fixtures/update_record.json")
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

	updatedRecord, err := client.UpdateTxtRRSet("example.dedyn.io", "_acme-challenge", []string{`"updated"`})
	require.NoError(t, err)

	expected := &RRSet{
		Name:    "_acme-challenge.example.dedyn.io.",
		Domain:  "example.dedyn.io",
		SubName: "_acme-challenge",
		Type:    "TXT",
		Records: []string{`"updated"`},
		TTL:     300,
	}
	assert.Equal(t, expected, updatedRecord)
}

func TestDeleteTxtRRSet(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("token")
	client.BaseURL = server.URL

	mux.HandleFunc("/domains/example.dedyn.io/rrsets/_acme-challenge/TXT/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteTxtRRSet("example.dedyn.io", "_acme-challenge")
	require.NoError(t, err)
}
