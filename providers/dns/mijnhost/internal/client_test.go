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

const apiKey = "secret"

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(apiKey)
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func testHandler(filename string, method string, statusCode int) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get(authorizationHeader)
		if auth != apiKey {
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

func TestClient_ListDomains(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains", testHandler("./list-domains.json", http.MethodGet, http.StatusOK))

	domains, err := client.ListDomains(context.Background())
	require.NoError(t, err)

	expected := []Domain{{
		ID:          1000,
		Domain:      "example.com",
		RenewalDate: "2030-01-01",
		Status:      "Active",
		StatusID:    1,
		Tags:        []string{"my-tag"},
	}}

	assert.Equal(t, expected, domains)
}

func TestClient_GetRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/example.com/dns", testHandler("./get-dns-records.json", http.MethodGet, http.StatusOK))

	records, err := client.GetRecords(context.Background(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			Type:  "A",
			Name:  "example.com.",
			Value: "135.226.123.12",
			TTL:   900,
		},
		{
			Type:  "AAAA",
			Name:  "example.com.",
			Value: "2009:21d0:322:6100::5:c92b",
			TTL:   900,
		},
		{
			Type:  "MX",
			Name:  "example.com.",
			Value: "10 mail.example.com.",
			TTL:   900,
		},
		{
			Type:  "TXT",
			Name:  "example.com.",
			Value: "v=spf1 include:spf.mijn.host ~all",
			TTL:   900,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_UpdateRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/domains/example.com/dns", testHandler("./update-dns-records.json", http.MethodPut, http.StatusOK))

	err := client.UpdateRecords(context.Background(), "example.com", nil)
	require.NoError(t, err)
}
