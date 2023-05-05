package internal

import (
	"context"
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
	const apiKey = "myKeyA"

	client, mux := setupTest(t, apiKey)

	const domain = "example.com"

	mux.HandleFunc("/cdn/4.0/domains/"+domain+"/dns-records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authorizationHeader)
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

	_, err := client.GetTxtRecord(context.Background(), domain, "_acme-challenge", "txtxtxt")
	require.NoError(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	const apiKey = "myKeyB"

	client, mux := setupTest(t, apiKey)

	const domain = "example.com"

	mux.HandleFunc("/cdn/4.0/domains/"+domain+"/dns-records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authorizationHeader)
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

	record := DNSRecord{
		Name:  "_acme-challenge",
		Type:  "txt",
		Value: &TXTRecordValue{Text: "txtxtxt"},
		TTL:   600,
	}

	newRecord, err := client.CreateRecord(context.Background(), domain, record)
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
	const apiKey = "myKeyC"

	client, mux := setupTest(t, apiKey)

	const domain = "example.com"
	const recordID = "recordId"

	mux.HandleFunc("/cdn/4.0/domains/"+domain+"/dns-records/"+recordID, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authorizationHeader)
		if auth != apiKey {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
			return
		}
	})

	err := client.DeleteRecord(context.Background(), domain, recordID)
	require.NoError(t, err)
}
