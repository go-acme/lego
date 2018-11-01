package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/api"
	"github.com/xenolf/lego/emca/certificate/certcrypto"
	"github.com/xenolf/lego/emca/le"
)

func TestCertifier_createOrderForIdentifiers(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			w.Header().Add("Replay-Nonce", "12345")
			w.Header().Add("Retry-After", "0")

			err := writeJSONResponse(w, le.Directory{
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
		case http.MethodPost:
			err := writeJSONResponse(w, le.OrderMessage{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	keyBits := 512 // small value keeps test fast

	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", ts.URL, "", key)
	require.NoError(t, err)

	resolver := &mockResolver{}

	certifier := NewCertifier(core, certcrypto.RSA2048, resolver)

	_, err = certifier.createOrderForIdentifiers([]string{"example.com"})
	require.NoError(t, err)
}

// writeJSONResponse marshals the body as JSON and writes it to the response.
func writeJSONResponse(w http.ResponseWriter, body interface{}) error {
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

type mockResolver struct{}

func (*mockResolver) Solve(authorizations []le.Authorization) error {
	return nil
}
