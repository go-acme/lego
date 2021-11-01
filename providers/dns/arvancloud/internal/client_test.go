package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetTxtRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	const domain = "example.com"
	const apiKey = "myKeyA"

	mux.HandleFunc("/cdn/4.0/domains/"+domain+"/dns-records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authHeader)
		if auth != apiKey {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
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

	client := NewClient(apiKey)
	client.BaseURL = server.URL

	_, err := client.GetTxtRecord(domain, "_acme-challenge", "txtxtxt")
	require.NoError(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	const domain = "example.com"
	const apiKey = "myKeyB"

	mux.HandleFunc("/cdn/4.0/domains/"+domain+"/dns-records", func(rw http.ResponseWriter, req *http.Request) {
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

		rw.WriteHeader(http.StatusCreated)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client := NewClient(apiKey)
	client.BaseURL = server.URL

	record := DNSRecord{
		Name:  "_acme-challenge",
		Type:  "txt",
		Value: &TXTRecordValue{Text: "txtxtxt"},
		TTL:   600,
	}

	newRecord, err := client.CreateRecord(domain, record)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:            "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Type:          "txt",
		Value:         map[string]interface{}{"text": "txtxtxt"},
		Name:          "_acme-challenge",
		TTL:           120,
		UpstreamHTTPS: "default",
		IPFilterMode: &IPFilterMode{
			Count:     "single",
			Order:     "none",
			GeoFilter: "none",
		},
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	const domain = "example.com"
	const apiKey = "myKeyC"
	const recordID = "recordId"

	mux.HandleFunc("/cdn/4.0/domains/"+domain+"/dns-records/"+recordID, func(rw http.ResponseWriter, req *http.Request) {
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

	client := NewClient(apiKey)
	client.BaseURL = server.URL

	err := client.DeleteRecord(domain, recordID)
	require.NoError(t, err)
}
