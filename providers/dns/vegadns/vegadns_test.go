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
	"github.com/xenolf/lego/platform/tester"
)

const testDomain = "example.com"

const ipPort = "127.0.0.1:2112"

var envTest = tester.NewEnvTest("SECRET_VEGADNS_KEY", "SECRET_VEGADNS_SECRET", "VEGADNS_URL")

type muxCallback func() *http.ServeMux

func TestNewDNSProvider_Fail(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	_, err := NewDNSProvider()
	assert.Error(t, err, "VEGADNS_URL env missing")
}

func TestDNSProvider_TimeoutSuccess(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	ts, err := startTestServer(muxSuccess)
	require.NoError(t, err)

	defer ts.Close()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	timeout, interval := provider.Timeout()
	assert.Equal(t, timeout, time.Duration(720000000000))
	assert.Equal(t, interval, time.Duration(60000000000))
}

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		desc          string
		callback      muxCallback
		expectedError string
	}{
		{
			desc:     "Success",
			callback: muxSuccess,
		},
		{
			desc:          "FailToFindZone",
			callback:      muxFailToFindZone,
			expectedError: "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in Present: Unable to find auth zone for fqdn _acme-challenge.example.com",
		},
		{
			desc:          "FailToCreateTXT",
			callback:      muxFailToCreateTXT,
			expectedError: "vegadns: Got bad answer from VegaDNS on CreateTXT. Code: 400. Message: ",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			ts, err := startTestServer(test.callback)
			require.NoError(t, err)

			defer ts.Close()

			provider, err := NewDNSProvider()
			require.NoError(t, err)

			err = provider.Present(testDomain, "token", "keyAuth")
			if len(test.expectedError) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	testCases := []struct {
		desc          string
		callback      muxCallback
		expectedError string
	}{
		{
			desc:     "Success",
			callback: muxSuccess,
		},
		{
			desc:          "FailToFindZone",
			callback:      muxFailToFindZone,
			expectedError: "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in CleanUp: Unable to find auth zone for fqdn _acme-challenge.example.com",
		},
		{
			desc:          "FailToGetRecordID",
			callback:      muxFailToGetRecordID,
			expectedError: "vegadns: couldn't get Record ID in CleanUp: Got bad answer from VegaDNS on GetRecordID. Code: 404. Message: ",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			ts, err := startTestServer(test.callback)
			require.NoError(t, err)

			defer ts.Close()

			provider, err := NewDNSProvider()
			require.NoError(t, err)

			err = provider.CleanUp(testDomain, "token", "keyAuth")
			if len(test.expectedError) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func muxSuccess() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") == "example.com" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, domainsResponseMock)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/1.0/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("domain_id") == "1" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, recordsResponseMock)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, recordCreatedResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/records/3", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, recordDeletedResponseMock)
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

func muxFailToFindZone() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return mux
}

func muxFailToCreateTXT() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") == testDomain {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, domainsResponseMock)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/1.0/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("domain_id") == "1" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, recordsResponseMock)
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

func muxFailToGetRecordID() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/1.0/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenResponseMock)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc("/1.0/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") == testDomain {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, domainsResponseMock)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/1.0/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
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
