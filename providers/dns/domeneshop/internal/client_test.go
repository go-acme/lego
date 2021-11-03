package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("token", "secret")

	client.baseURL = server.URL

	return client, mux
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/domains/1/dns", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic dG9rZW46c2VjcmV0" {
			http.Error(rw, "invalid method: "+req.Method, http.StatusUnauthorized)
			return
		}

		_, _ = rw.Write([]byte(`{"id": 1}`))
	})

	err := client.CreateTXTRecord(&Domain{ID: 1}, "example", "txtTXTtxt")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/domains/1/dns", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic dG9rZW46c2VjcmV0" {
			http.Error(rw, "invalid method: "+req.Method, http.StatusUnauthorized)
			return
		}

		_, _ = rw.Write([]byte(`[
  {
    "id": 1,
    "host": "example.com",
    "ttl": 3600,
    "type": "TXT",
    "data": "txtTXTtxt"
  }
]`))
	})

	mux.HandleFunc("/domains/1/dns/1", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic dG9rZW46c2VjcmV0" {
			http.Error(rw, "invalid method: "+req.Method, http.StatusUnauthorized)
			return
		}
	})

	err := client.DeleteTXTRecord(&Domain{ID: 1}, "example.com", "txtTXTtxt")
	require.NoError(t, err)
}

func TestClient_getDNSRecordByHostData(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/domains/1/dns", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic dG9rZW46c2VjcmV0" {
			http.Error(rw, "invalid method: "+req.Method, http.StatusUnauthorized)
			return
		}

		_, _ = rw.Write([]byte(`[
  {
    "id": 1,
    "host": "example.com",
    "ttl": 3600,
    "type": "TXT",
    "data": "txtTXTtxt"
  }
]`))
	})

	record, err := client.getDNSRecordByHostData(Domain{ID: 1}, "example.com", "txtTXTtxt")
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:   1,
		Type: "TXT",
		Host: "example.com",
		Data: "txtTXTtxt",
		TTL:  3600,
	}

	assert.Equal(t, expected, record)
}

func TestClient_GetDomainByName(t *testing.T) {
	client, mux := setup(t)

	mux.HandleFunc("/domains", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("Authorization")
		if auth != "Basic dG9rZW46c2VjcmV0" {
			http.Error(rw, "invalid method: "+req.Method, http.StatusUnauthorized)
			return
		}

		_, _ = rw.Write([]byte(`[
  {
    "id": 1,
    "domain": "example.com",
    "expiry_date": "2019-08-24",
    "registered_date": "2019-08-24",
    "renew": true,
    "registrant": "Ola Nordmann",
    "status": "active",
    "nameservers": [
      "ns1.hyp.net",
      "ns2.hyp.net",
      "ns3.hyp.net"
    ],
    "services": {
      "registrar": true,
      "dns": true,
      "email": true,
      "webhotel": "none"
    }
  }
]`))
	})

	domain, err := client.GetDomainByName("example.com")
	require.NoError(t, err)

	expected := &Domain{
		Name:           "example.com",
		ID:             1,
		ExpiryDate:     "2019-08-24",
		Nameservers:    []string{"ns1.hyp.net", "ns2.hyp.net", "ns3.hyp.net"},
		RegisteredDate: "2019-08-24",
		Registrant:     "Ola Nordmann",
		Renew:          true,
		Services: Service{
			DNS:       true,
			Email:     true,
			Registrar: true,
			Webhotel:  "none",
		},
		Status: "active",
	}

	assert.Equal(t, expected, domain)
}
