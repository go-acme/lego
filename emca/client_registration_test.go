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
	keyBits := 512

	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	require.NoError(t, err, "Could not generate test key")

	user := mockUser{
		email:      "test@test.com",
		regres:     new(le.RegistrationResource),
		privatekey: key,
	}

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/directory":
			err = writeJSONResponse(w, le.Directory{
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
		case "/nonce":
			w.Header().Add("Replay-Nonce", "12345")
			w.Header().Add("Retry-After", "0")
		case "/account":
			w.Header().Set("Location", ts.URL+"/account_recovery")
		case "/account_recovery":
			err = writeJSONResponse(w, le.AccountMessage{
				Status: "valid",
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}))

	config := NewDefaultConfig(user).WithCADirURL(ts.URL + "/directory")

	client, err := NewClient(config)
	require.NoError(t, err, "Could not create client")

	res, err := client.ResolveAccountByKey()
	require.NoError(t, err, "Unexpected error resolving account by key")

	assert.Equal(t, "valid", res.Body.Status, "Unexpected account status")
}
