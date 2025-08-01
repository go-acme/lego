package tester

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
)

// MockACMEServer Minimal stub ACME server for validation.
func MockACMEServer() *servermock.Builder[string] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (string, error) {
			return server.URL, nil
		}).
		Route("GET /dir", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			serverURL := fmt.Sprintf("https://%s", req.Context().Value(http.LocalAddrContextKey))

			servermock.JSONEncode(acme.Directory{
				NewNonceURL:   serverURL + "/nonce",
				NewAccountURL: serverURL + "/account",
				NewOrderURL:   serverURL + "/newOrder",
				RevokeCertURL: serverURL + "/revokeCert",
				KeyChangeURL:  serverURL + "/keyChange",
				RenewalInfo:   serverURL + "/renewalInfo",
			}).ServeHTTP(rw, req)
		})).
		Route("HEAD /nonce", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Replay-Nonce", "12345")
			rw.Header().Set("Retry-After", "0")
		}))
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
