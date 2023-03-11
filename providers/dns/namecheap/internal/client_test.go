package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", handler)

	client := NewClient("user", "secret", "127.0.0.1")
	client.HTTPClient = server.Client()
	client.BaseURL = server.URL

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

func TestClient_GetHosts(t *testing.T) {
	client := setupTest(t, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		expectedParams := map[string]string{
			"ApiKey":   "secret",
			"ApiUser":  "user",
			"ClientIp": "127.0.0.1",
			"Command":  "namecheap.domains.dns.getHosts",
			"SLD":      "foo",
			"TLD":      "example.com",
			"UserName": "user",
		}

		query := req.URL.Query()
		for k, v := range expectedParams {
			if query.Get(k) != v {
				http.Error(rw, fmt.Sprintf("invalid query parameter %s value: %s", k, query.Get(k)), http.StatusBadRequest)
				return
			}
		}

		writeFixture(rw, "getHosts.xml")
	})

	hosts, err := client.GetHosts(context.Background(), "foo", "example.com")
	require.NoError(t, err)

	expected := []Record{
		{Type: "A", Name: "@", Address: "1.2.3.4", MXPref: "10", TTL: "1800"},
		{Type: "A", Name: "www", Address: "122.23.3.7", MXPref: "10", TTL: "1800"},
	}

	assert.Equal(t, expected, hosts)
}

func TestClient_GetHosts_error(t *testing.T) {
	client := setupTest(t, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		writeFixture(rw, "getHosts_errorBadAPIKey1.xml")
	})

	_, err := client.GetHosts(context.Background(), "foo", "example.com")
	require.ErrorAs(t, err, &apiError{})
}

func TestClient_SetHosts(t *testing.T) {
	client := setupTest(t, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			http.Error(rw, fmt.Sprintf("invalid Content-Type: %s", req.Header.Get("Content-Type")), http.StatusBadRequest)
			return
		}

		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		expectedParams := map[string]string{
			"HostName1":   "_acme-challenge.test.example.com",
			"RecordType1": "TXT",
			"Address1":    "txtTXTtxt",
			"MXPref1":     "10",
			"TTL1":        "120",

			"HostName2":   "_acme-challenge.test.example.org",
			"RecordType2": "TXT",
			"Address2":    "txtTXTtxt",
			"MXPref2":     "10",
			"TTL2":        "120",

			"ApiKey":   "secret",
			"ApiUser":  "user",
			"ClientIp": "127.0.0.1",
			"Command":  "namecheap.domains.dns.setHosts",
			"SLD":      "foo",
			"TLD":      "example.com",
			"UserName": "user",
		}

		for k, v := range expectedParams {
			if req.Form.Get(k) != v {
				http.Error(rw, fmt.Sprintf("invalid form data %s value: %q", k, req.Form.Get(k)), http.StatusBadRequest)
				return
			}
		}

		writeFixture(rw, "setHosts.xml")
	})

	records := []Record{
		{Name: "_acme-challenge.test.example.com", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
		{Name: "_acme-challenge.test.example.org", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
	}

	err := client.SetHosts(context.Background(), "foo", "example.com", records)
	require.NoError(t, err)
}

func TestClient_SetHosts_error(t *testing.T) {
	client := setupTest(t, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		writeFixture(rw, "setHosts_errorBadAPIKey1.xml")
	})

	records := []Record{
		{Name: "_acme-challenge.test.example.com", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
		{Name: "_acme-challenge.test.example.org", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
	}

	err := client.SetHosts(context.Background(), "foo", "example.com", records)
	require.ErrorAs(t, err, &apiError{})
}
