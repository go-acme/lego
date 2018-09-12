// Package vegadns implements a DNS provider for solving the DNS-01
// challenge using VegaDNS.
package vegadns

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ipPort = "127.0.0.1:2112"

var jsonMap = map[string]string{
	"token": `
{
  "access_token":"699dd4ff-e381-46b8-8bf8-5de49dd56c1f",
  "token_type":"bearer",
  "expires_in":3600
}
`,
	"domains": `
{
  "domains":[
    {
      "domain_id":1,
      "domain":"example.com",
      "status":"active",
      "owner_id":0
    }
  ]
}
`,
	"records": `
{
  "status":"ok",
  "total_records":2,
  "domain":{
    "status":"active",
    "domain":"example.com",
    "owner_id":0,
    "domain_id":1
  },
  "records":[
    {
      "retry":"2048",
      "minimum":"2560",
      "refresh":"16384",
      "email":"hostmaster.example.com",
      "record_type":"SOA",
      "expire":"1048576",
      "ttl":86400,
      "record_id":1,
      "nameserver":"ns1.example.com",
      "domain_id":1,
      "serial":""
    },
    {
      "name":"example.com",
      "value":"ns1.example.com",
      "record_type":"NS",
      "ttl":3600,
      "record_id":2,
      "location_id":null,
      "domain_id":1
    },
    {
      "name":"_acme-challenge.example.com",
      "value":"my_challenge",
      "record_type":"TXT",
      "ttl":3600,
      "record_id":3,
      "location_id":null,
      "domain_id":1
    }
  ]
}
`,
	"recordCreated": `
{
  "status":"ok",
  "record":{
    "name":"_acme-challenge.example.com",
    "value":"my_challenge",
    "record_type":"TXT",
    "ttl":3600,
    "record_id":3,
    "location_id":null,
    "domain_id":1
  }
}
`,
	"recordDeleted": `{"status": "ok"}`,
}

type muxCallback func() *http.ServeMux

func TestVegaDNSNewDNSProviderFail(t *testing.T) {
	os.Setenv("VEGADNS_URL", "")
	_, err := NewDNSProvider()
	assert.Error(t, err, "VEGADNS_URL env missing")
}

func TestVegaDNSTimeoutSuccess(t *testing.T) {
	ts, err := startTestServer(vegaDNSMuxSuccess)
	require.NoError(t, err)

	defer ts.Close()
	defer os.Clearenv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	timeout, interval := provider.Timeout()
	assert.Equal(t, timeout, time.Duration(720000000000))
	assert.Equal(t, interval, time.Duration(60000000000))
}

func TestVegaDNSPresentSuccess(t *testing.T) {
	ts, err := startTestServer(vegaDNSMuxSuccess)
	require.NoError(t, err)

	defer ts.Close()
	defer os.Clearenv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	assert.NoError(t, err)
}

func TestVegaDNSPresentFailToFindZone(t *testing.T) {
	ts, err := startTestServer(vegaDNSMuxFailToFindZone)
	require.NoError(t, err)

	defer ts.Close()
	defer os.Clearenv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	assert.EqualError(t, err, "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in Present: Unable to find auth zone for fqdn _acme-challenge.example.com")
}

func TestVegaDNSPresentFailToCreateTXT(t *testing.T) {
	ts, err := startTestServer(vegaDNSMuxFailToCreateTXT)
	require.NoError(t, err)

	defer ts.Close()
	defer os.Clearenv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present("example.com", "token", "keyAuth")
	assert.EqualError(t, err, "vegadns: Got bad answer from VegaDNS on CreateTXT. Code: 400. Message: ")
}

func TestVegaDNSCleanUpSuccess(t *testing.T) {
	ts, err := startTestServer(vegaDNSMuxSuccess)
	require.NoError(t, err)

	defer ts.Close()
	defer os.Clearenv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp("example.com", "token", "keyAuth")
	assert.NoError(t, err)
}

func TestVegaDNSCleanUpFailToFindZone(t *testing.T) {
	ts, err := startTestServer(vegaDNSMuxFailToFindZone)
	require.NoError(t, err)

	defer ts.Close()
	defer os.Clearenv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp("example.com", "token", "keyAuth")
	assert.EqualError(t, err, "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in CleanUp: Unable to find auth zone for fqdn _acme-challenge.example.com")
}

func TestVegaDNSCleanUpFailToGetRecordID(t *testing.T) {
	ts, err := startTestServer(vegaDNSMuxFailToGetRecordID)
	require.NoError(t, err)

	defer ts.Close()
	defer os.Clearenv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp("example.com", "token", "keyAuth")
	assert.EqualError(t, err, "vegadns: couldn't get Record ID in CleanUp: Got bad answer from VegaDNS on GetRecordID. Code: 404. Message: ")
}

func vegaDNSMuxSuccess() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["token"])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") == "example.com" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["domains"])
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/1.0/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("domain_id") == "1" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, jsonMap["records"])
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, jsonMap["recordCreated"])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/records/3", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["recordDeleted"])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Not Found for Request: (%+v)\n\n", r)
	})

	return mux
}

func vegaDNSMuxFailToFindZone() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["token"])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return mux
}

func vegaDNSMuxFailToCreateTXT() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["token"])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") == "example.com" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["domains"])
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/1.0/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("domain_id") == "1" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, jsonMap["records"])
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		case http.MethodPost:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	return mux
}

func vegaDNSMuxFailToGetRecordID() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["token"])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") == "example.com" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, jsonMap["domains"])
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/1.0/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	return mux
}

// Starts and returns a test server using a custom ip/port. Defer close() afterwards.
func startTestServer(callback muxCallback) (*httptest.Server, error) {
	err := os.Setenv("SECRET_VEGADNS_KEY", "key")
	if err != nil {
		return nil, err
	}

	err = os.Setenv("SECRET_VEGADNS_SECRET", "secret")
	if err != nil {
		return nil, err
	}

	err = os.Setenv("VEGADNS_URL", "http://"+ipPort)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewUnstartedServer(callback())

	l, err := net.Listen("tcp", ipPort)
	if err != nil {
		return nil, err
	}

	ts.Listener = l
	ts.Start()

	return ts, nil
}
