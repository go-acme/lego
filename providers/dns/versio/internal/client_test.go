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

func setupTest(t *testing.T, pattern string, h http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, h)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.BaseURL, _ = url.Parse(server.URL)

	return client
}

func writeFixture(rw http.ResponseWriter, filename string) {
	file, err := os.Open(filepath.Join("fixtures", filename))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	_, _ = io.Copy(rw, file)
}

func TestClient_GetDomain(t *testing.T) {
	client := setupTest(t, "/domains/example.com", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic dXNlcjpzZWNyZXQ=" {
			http.Error(rw, "invalid credentials: "+auth, http.StatusUnauthorized)
			return
		}

		writeFixture(rw, "get-domain.json")
	})

	records, err := client.GetDomain(context.Background(), "example.com")
	require.NoError(t, err)

	expected := &DomainInfoResponse{DomainInfo: DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}}

	assert.Equal(t, expected, records)
}

func TestClient_GetDomain_error(t *testing.T) {
	client := setupTest(t, "/domains/example.com", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		rw.WriteHeader(http.StatusUnauthorized)

		writeFixture(rw, "get-domain-error.json")
	})

	_, err := client.GetDomain(context.Background(), "example.com")
	require.ErrorAs(t, err, &ErrorMessage{})
}

func TestClient_UpdateDomain(t *testing.T) {
	client := setupTest(t, "/domains/example.com/update", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic dXNlcjpzZWNyZXQ=" {
			http.Error(rw, "invalid credentials: "+auth, http.StatusUnauthorized)
			return
		}

		writeFixture(rw, "update-domain.json")
	})

	msg := &DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}

	records, err := client.UpdateDomain(context.Background(), "example.com", msg)
	require.NoError(t, err)

	expected := &DomainInfoResponse{DomainInfo: DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}}

	assert.Equal(t, expected, records)
}

func TestClient_UpdateDomain_error(t *testing.T) {
	client := setupTest(t, "/domains/example.com/update", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		rw.WriteHeader(http.StatusUnauthorized)

		writeFixture(rw, "update-domain.json")
	})

	msg := &DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}

	_, err := client.UpdateDomain(context.Background(), "example.com", msg)
	require.ErrorAs(t, err, &ErrorMessage{})
}
