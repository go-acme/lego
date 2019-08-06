package registration

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v3/acme"
	"github.com/go-acme/lego/v3/acme/api"
	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrar_ResolveAccountByKey(t *testing.T) {
	mux, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	mux.HandleFunc("/account", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", apiURL+"/account_recovery")
		_, err := w.Write([]byte("{}"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/account_recovery", func(w http.ResponseWriter, _ *http.Request) {
		err := tester.WriteJSONResponse(w, acme.Account{
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

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	registrar := NewRegistrar(core, user)

	res, err := registrar.ResolveAccountByKey()
	require.NoError(t, err, "Unexpected error resolving account by key")

	assert.Equal(t, "valid", res.Body.Status, "Unexpected account status")
}
