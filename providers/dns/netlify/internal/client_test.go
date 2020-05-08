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

func TestClient_GetRecords(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/dns_zones/zoneID/dns_records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "unsupported method", http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer tokenA" {
			http.Error(rw, fmt.Sprintf("invali token: %s", auth), http.StatusUnauthorized)
			return
		}

		rw.Header().Set("Content-Type", "application/json; charset=utf-8")

		file, err := os.Open("./fixtures/get_records.json")
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

	client := NewClient("tokenA")
	client.BaseURL = server.URL

	records, err := client.GetRecords("zoneID")
	require.NoError(t, err)

	expected := []DNSRecord{
		{ID: "u6b433c15a27a2d79c6616d6", Hostname: "example.org", TTL: 3600, Type: "A", Value: "10.10.10.10"},
		{ID: "u6b4764216f272872ac0ff71", Hostname: "test.example.org", TTL: 300, Type: "TXT", Value: "txtxtxtxtxtxt"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/dns_zones/zoneID/dns_records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "unsupported method", http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer tokenB" {
			http.Error(rw, fmt.Sprintf("invali token: %s", auth), http.StatusUnauthorized)
			return
		}

		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(http.StatusCreated)

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

	client := NewClient("tokenB")
	client.BaseURL = server.URL

	record := DNSRecord{
		Hostname: "_acme-challenge.example.com",
		TTL:      300,
		Type:     "TXT",
		Value:    "txtxtxtxtxtxt",
	}

	result, err := client.CreateRecord("zoneID", record)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:       "u6b4764216f272872ac0ff71",
		Hostname: "test.example.org",
		TTL:      300,
		Type:     "TXT",
		Value:    "txtxtxtxtxtxt",
	}

	assert.Equal(t, expected, result)
}

func TestClient_RemoveRecord(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/dns_zones/zoneID/dns_records/recordID", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "unsupported method", http.StatusMethodNotAllowed)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer tokenC" {
			http.Error(rw, fmt.Sprintf("invali token: %s", auth), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusNoContent)
	})

	client := NewClient("tokenC")
	client.BaseURL = server.URL

	err := client.RemoveRecord("zoneID", "recordID")
	require.NoError(t, err)
}
