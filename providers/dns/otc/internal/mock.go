package internal

import (
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

const fakeOTCToken = "62244bc21da68d03ebac94e6636ff01f"

func writeFixture(rw http.ResponseWriter, filename string) {
	file, err := os.Open(filepath.Join("internal", "fixtures", filename))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	_, _ = io.Copy(rw, file)
}

// DNSServerMock mock.
type DNSServerMock struct {
	t      *testing.T
	server *httptest.Server
	mux    *http.ServeMux
}

// NewDNSServerMock create a new DNSServerMock.
func NewDNSServerMock(t *testing.T) *DNSServerMock {
	t.Helper()

	mux := http.NewServeMux()

	return &DNSServerMock{
		t:      t,
		server: httptest.NewServer(mux),
		mux:    mux,
	}
}

func (m *DNSServerMock) GetServerURL() string {
	return m.server.URL
}

// ShutdownServer creates the mock server.
func (m *DNSServerMock) ShutdownServer() {
	m.server.Close()
}

// HandleAuthSuccessfully Handle auth successfully.
func (m *DNSServerMock) HandleAuthSuccessfully() {
	m.mux.HandleFunc("/v3/auth/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Subject-Token", fakeOTCToken)

		_, _ = fmt.Fprintf(w, `{
		  "token": {
		    "catalog": [
		      {
			"type": "dns",
			"id": "56cd81db1f8445d98652479afe07c5ba",
			"name": "",
			"endpoints": [
			  {
			    "url": "%s",
			    "region": "eu-de",
			    "region_id": "eu-de",
			    "interface": "public",
			    "id": "0047a06690484d86afe04877074efddf"
			  }
			]
		      }
		    ]
		  }}`, m.server.URL)
	})
}

// HandleListZonesSuccessfully Handle list zones successfully.
func (m *DNSServerMock) HandleListZonesSuccessfully() {
	m.mux.HandleFunc("/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(m.t, http.MethodGet, r.Method)
		assert.Equal(m.t, "/v2/zones", r.URL.Path)
		assert.Equal(m.t, "name=example.com.", r.URL.RawQuery)
		assert.Equal(m.t, "application/json", r.Header.Get("Accept"))

		writeFixture(w, "zones_GET.json")
	})
}

// HandleListZonesEmpty Handle list zones empty.
func (m *DNSServerMock) HandleListZonesEmpty() {
	m.mux.HandleFunc("/v2/zones", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(m.t, http.MethodGet, r.Method)
		assert.Equal(m.t, "/v2/zones", r.URL.Path)
		assert.Equal(m.t, "name=example.com.", r.URL.RawQuery)
		assert.Equal(m.t, "application/json", r.Header.Get("Accept"))

		writeFixture(w, "zones_GET_empty.json")
	})
}

// HandleDeleteRecordsetsSuccessfully Handle delete recordsets successfully.
func (m *DNSServerMock) HandleDeleteRecordsetsSuccessfully() {
	m.mux.HandleFunc("/v2/zones/123123/recordsets/321321", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(m.t, http.MethodDelete, r.Method)
		assert.Equal(m.t, "/v2/zones/123123/recordsets/321321", r.URL.Path)
		assert.Equal(m.t, "application/json", r.Header.Get("Accept"))

		writeFixture(w, "zones-recordsets_DELETE.json")
	})
}

// HandleListRecordsetsEmpty Handle list recordsets empty.
func (m *DNSServerMock) HandleListRecordsetsEmpty() {
	m.mux.HandleFunc("/v2/zones/123123/recordsets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(m.t, "/v2/zones/123123/recordsets", r.URL.Path)
		assert.Equal(m.t, "name=_acme-challenge.example.com.&type=TXT", r.URL.RawQuery)

		writeFixture(w, "zones-recordsets_GET_empty.json")
	})
}

// HandleListRecordsetsSuccessfully Handle list recordsets successfully.
func (m *DNSServerMock) HandleListRecordsetsSuccessfully() {
	m.mux.HandleFunc("/v2/zones/123123/recordsets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(m.t, "application/json", r.Header.Get("Accept"))

		if r.Method == http.MethodGet {
			assert.Equal(m.t, "/v2/zones/123123/recordsets", r.URL.Path)
			assert.Equal(m.t, "name=_acme-challenge.example.com.&type=TXT", r.URL.RawQuery)

			writeFixture(w, "zones-recordsets_GET.json")
			return
		}

		if r.Method == http.MethodPost {
			assert.Equal(m.t, "application/json", r.Header.Get("Content-Type"))

			raw, err := io.ReadAll(r.Body)
			require.NoError(m.t, err)
			exceptedString := `{
				"name": "_acme-challenge.example.com.",
				"description": "Added TXT record for ACME dns-01 challenge using lego client",
				"type": "TXT",
				"ttl": 300,
				"records": ["\"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI\""]
			}`

			assert.JSONEq(m.t, exceptedString, string(raw))

			writeFixture(w, "zones-recordsets_POST.json")
			return
		}

		http.Error(w, fmt.Sprintf("Expected method to be 'GET' or 'POST' but got '%s'", r.Method), http.StatusBadRequest)
	})
}
