package tester

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/xenolf/lego/le"
)

// SetupFakeAPI Minimal stub ACME server for validation.
func SetupFakeAPI() (*http.ServeMux, string, func()) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		err := WriteJSONResponse(w, le.Directory{
			NewNonceURL:   ts.URL + "/nonce",
			NewAccountURL: ts.URL + "/account",
			NewOrderURL:   ts.URL + "/newOrder",
			RevokeCertURL: ts.URL + "/revokeCert",
			KeyChangeURL:  ts.URL + "/keyChange",
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return mux, ts.URL, ts.Close
}

// WriteJSONResponse marshals the body as JSON and writes it to the response.
func WriteJSONResponse(w http.ResponseWriter, body interface{}) error {
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
