package internal

import (
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

func setupTest(t *testing.T, apiToken string) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("me", apiToken)
	client.baseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, mux
}

func TestClient_ListZones(t *testing.T) {
	client, mux := setupTest(t, "secretA")

	mux.HandleFunc("/v1/zones", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("invalid method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic bWU6c2VjcmV0QQ==" {
			http.Error(rw, fmt.Sprintf("invalid authentication: %s", auth), http.StatusUnauthorized)
			return
		}

		file, err := os.Open("./fixtures/list_zones.json")
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

	zones, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []DNSZone{
		{
			ID:             1,
			Name:           "example.com",
			Synced:         false,
			QueriesCount:   0,
			RecordsCount:   3,
			AliasesCount:   0,
			RedirectsCount: 0,
			ForwardsCount:  0,
			TemplateID:     0,
		},
		{
			ID:             2,
			Name:           "example.net",
			Synced:         false,
			QueriesCount:   0,
			RecordsCount:   3,
			AliasesCount:   0,
			RedirectsCount: 0,
			ForwardsCount:  0,
			TemplateID:     0,
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_CreateRecord(t *testing.T) {
	client, mux := setupTest(t, "secretB")

	mux.HandleFunc("/v1/zones/1/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("invalid method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic bWU6c2VjcmV0Qg==" {
			http.Error(rw, fmt.Sprintf("invalid authentication: %s", auth), http.StatusUnauthorized)
			return
		}

		file, err := os.Open("./fixtures/create_record.json")
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

	zone := DNSZone{ID: 1}

	record := DNSRecord{
		Name:    "example.com.",
		Type:    "MX",
		Content: "10 mail.example.com.",
		TTL:     300,
	}

	newRecord, err := client.CreateRecord(t.Context(), zone, record)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:      100,
		Name:    "example.com.",
		Type:    "MX",
		Content: "10 mail.example.com.",
		TTL:     300,
		ZoneID:  1,
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t, "secretC")

	mux.HandleFunc("/v1/zones/1/records/2", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, fmt.Sprintf("invalid method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic bWU6c2VjcmV0Qw==" {
			http.Error(rw, fmt.Sprintf("invalid authentication: %s", auth), http.StatusUnauthorized)
			return
		}

		file, err := os.Open("./fixtures/delete_record.json")
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

	record := &DNSRecord{
		ID:      2,
		Name:    "example.com.",
		Type:    "MX",
		Content: "10 mail.example.com.",
		TTL:     300,
		ZoneID:  1,
	}

	err := client.DeleteRecord(t.Context(), record)
	require.NoError(t, err)
}
