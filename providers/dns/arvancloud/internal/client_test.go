package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetTxtRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	const domain = "example.com"
	const apiKey = "Apikey XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

	mux.HandleFunc(fmt.Sprintf("/domains/%s/dns-records", domain), func(rw http.ResponseWriter, req *http.Request) {
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

	record, err := client.GetTxtRecord(domain, "TEST_NAME", "TEST_VALUE")
	require.NoError(t, err)

	fmt.Println(record)
}

func TestClient_CreateRecord(t *testing.T) {
	mux := http.NewServeMux()

	const domain = "example.com"
	const apiKey = "Apikey XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

	mux.HandleFunc(fmt.Sprintf("/domains/%s/dns-records", domain), func(rw http.ResponseWriter, req *http.Request) {
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

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client := NewClient(apiKey)

	record := DNSRecord{
		Type:   "txt",
		Name:   "TEST_NAME",
		Value:  TxtValue{Text: "TEST_VALUE"},
		TTL:    120,
		UpstreamHTTPS: "default",
		IPFilterMode: IPFilterMode{
			Count: "single",
			GeoFilter: "none",
			Order: "none",
		},
	}

	err := client.CreateRecord(domain, record)
	require.NoError(t, err)
}


func TestClient_DeleteRecord(t *testing.T) {
	mux := http.NewServeMux()

	const domain = "example.com"
	const recordID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	const apiKey = "Apikey XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

	mux.HandleFunc(fmt.Sprintf("/domains/%s/dns-records/%s", domain, recordID), func(rw http.ResponseWriter, req *http.Request) {
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

	err := client.DeleteRecord(domain, recordID)
	require.NoError(t, err)
}
