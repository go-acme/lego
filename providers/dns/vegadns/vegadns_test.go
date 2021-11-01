package vegadns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDomain = "example.com"

var envTest = tester.NewEnvTest(EnvKey, EnvSecret, EnvURL)

func TestNewDNSProvider_Fail(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	_, err := NewDNSProvider()
	assert.Error(t, err, "VEGADNS_URL env missing")
}

func TestDNSProvider_TimeoutSuccess(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	setupTest(t, muxSuccess())

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	timeout, interval := provider.Timeout()
	assert.Equal(t, timeout, 12*time.Minute)
	assert.Equal(t, interval, 1*time.Minute)
}

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		desc          string
		handler       http.Handler
		expectedError string
	}{
		{
			desc:    "Success",
			handler: muxSuccess(),
		},
		{
			desc:          "FailToFindZone",
			handler:       muxFailToFindZone(),
			expectedError: "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in Present: Unable to find auth zone for fqdn _acme-challenge.example.com",
		},
		{
			desc:          "FailToCreateTXT",
			handler:       muxFailToCreateTXT(),
			expectedError: "vegadns: Got bad answer from VegaDNS on CreateTXT. Code: 400. Message: ",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			setupTest(t, test.handler)

			provider, err := NewDNSProvider()
			require.NoError(t, err)

			err = provider.Present(testDomain, "token", "keyAuth")
			if test.expectedError == "" {
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
		handler       http.Handler
		expectedError string
	}{
		{
			desc:    "Success",
			handler: muxSuccess(),
		},
		{
			desc:          "FailToFindZone",
			handler:       muxFailToFindZone(),
			expectedError: "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in CleanUp: Unable to find auth zone for fqdn _acme-challenge.example.com",
		},
		{
			desc:          "FailToGetRecordID",
			handler:       muxFailToGetRecordID(),
			expectedError: "vegadns: couldn't get Record ID in CleanUp: Got bad answer from VegaDNS on GetRecordID. Code: 404. Message: ",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			setupTest(t, test.handler)

			provider, err := NewDNSProvider()
			require.NoError(t, err)

			err = provider.CleanUp(testDomain, "token", "keyAuth")
			if test.expectedError == "" {
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

func setupTest(t *testing.T, mux http.Handler) {
	t.Helper()

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	envTest.Apply(map[string]string{
		EnvKey:    "key",
		EnvSecret: "secret",
		EnvURL:    server.URL,
	})
}
