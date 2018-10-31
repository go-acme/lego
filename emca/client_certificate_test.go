package emca

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/le"
)

func TestClient_createOrderForIdentifiers(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			w.Header().Add("Replay-Nonce", "12345")
			w.Header().Add("Retry-After", "0")
			err := writeJSONResponse(w, le.Directory{
				NewNonceURL:   ts.URL,
				NewAccountURL: ts.URL,
				NewOrderURL:   ts.URL,
				RevokeCertURL: ts.URL,
				KeyChangeURL:  ts.URL,
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

	user := mockUser{
		email:      "test@test.com",
		regres:     &le.RegistrationResource{URI: ts.URL},
		privatekey: key,
	}

	config := NewDefaultConfig(user).WithCADirURL(ts.URL)

	client, err := NewClient(config)
	require.NoError(t, err, "Could not create client")

	_, err = client.createOrderForIdentifiers([]string{"example.com"})
	require.NoError(t, err)
}
