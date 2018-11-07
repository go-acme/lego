package registration

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/le/api"
	"github.com/xenolf/lego/platform/tester"
)

func TestRegistrar_ResolveAccountByKey(t *testing.T) {
	mux, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	mux.HandleFunc("/nonce", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Header().Add("Retry-After", "0")
	})

	mux.HandleFunc("/account", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", apiURL+"/account_recovery")
		w.Write([]byte("{}"))
	})

	mux.HandleFunc("/account_recovery", func(w http.ResponseWriter, r *http.Request) {
		err := tester.WriteJSONResponse(w, le.Account{
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
		regres:     &Resource{},
		privatekey: key,
	}

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/directory", "", key)
	require.NoError(t, err)

	registrar := NewRegistrar(core, user)

	res, err := registrar.ResolveAccountByKey()
	require.NoError(t, err, "Unexpected error resolving account by key")

	assert.Equal(t, "valid", res.Body.Status, "Unexpected account status")
}
