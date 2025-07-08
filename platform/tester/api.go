package tester

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/acme"
)

// SetupFakeAPI Minimal stub ACME server for validation.
func SetupFakeAPI(t *testing.T) (*http.ServeMux, string) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/dir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		err := WriteJSONResponse(w, acme.Directory{
			NewNonceURL:   server.URL + "/nonce",
			NewAccountURL: server.URL + "/account",
			NewOrderURL:   server.URL + "/newOrder",
			RevokeCertURL: server.URL + "/revokeCert",
			KeyChangeURL:  server.URL + "/keyChange",
			RenewalInfo:   server.URL + "/renewalInfo",
		})

		mux.HandleFunc("/nonce", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodHead {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			w.Header().Set("Replay-Nonce", "12345")
			w.Header().Set("Retry-After", "0")
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return mux, server.URL
}

// WriteJSONResponse marshals the body as JSON and writes it to the response.
func WriteJSONResponse(w http.ResponseWriter, body any) error {
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bs); err != nil {
		return err
	}

	return nil
}
