package emca

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/le"
)

func TestClient_ResolveAccountByKey(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)

	mux.HandleFunc("/directory", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("/nonce", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Header().Add("Retry-After", "0")
	})

	mux.HandleFunc("/account", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", ts.URL+"/account_recovery")
	})

	mux.HandleFunc("/account_recovery", func(w http.ResponseWriter, r *http.Request) {
		err := writeJSONResponse(w, le.AccountMessage{
			Status: "valid",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	key, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Could not generate test key")

	user := mockUser{
		email:      "test@test.com",
		regres:     new(le.RegistrationResource),
		privatekey: key,
	}

	config := NewDefaultConfig(user).WithCADirURL(ts.URL + "/directory")

	client, err := NewClient(config)
	require.NoError(t, err, "Could not create client")

	res, err := client.ResolveAccountByKey()
	require.NoError(t, err, "Unexpected error resolving account by key")

	assert.Equal(t, "valid", res.Body.Status, "Unexpected account status")
}
