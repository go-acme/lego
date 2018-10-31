package emca

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/certificate/certcrypto"
	"github.com/xenolf/lego/emca/le"
)

func TestNewClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(le.Directory{
			NewNonceURL:   "http://test",
			NewAccountURL: "http://test",
			NewOrderURL:   "http://test",
			RevokeCertURL: "http://test",
			KeyChangeURL:  "http://test",
		})

		_, err := w.Write(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	require.NoError(t, err, "Could not generate test key")

	keyType := certcrypto.RSA2048

	user := mockUser{
		email:      "test@test.com",
		regres:     new(le.RegistrationResource),
		privatekey: key,
	}

	config := NewDefaultConfig(user).
		WithCADirURL(ts.URL).
		WithKeyType(keyType)

	client, err := NewClient(config)
	require.NoError(t, err, "Could not create client")

	require.NotNil(t, client.jws, "client.jws")
	assert.Equal(t, keyType, client.keyType, "client.keyType")
	assert.Len(t, client.solvers, 2, "solvers")
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
