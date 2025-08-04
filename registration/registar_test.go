package registration

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrar_ResolveAccountByKey(t *testing.T) {
	server := tester.MockACMEServer().
		Route("/account",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("Location",
					fmt.Sprintf("http://%s/account", req.Context().Value(http.LocalAddrContextKey)))

				servermock.JSONEncode(acme.Account{Status: "valid"}).ServeHTTP(rw, req)
			})).
		BuildHTTPS(t)

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err, "Could not generate test key")

	user := mockUser{
		email:      "test@test.com",
		regres:     &Resource{},
		privatekey: key,
	}

	core, err := api.New(server.Client(), "lego-test", server.URL+"/dir", "", key)
	require.NoError(t, err)

	registrar := NewRegistrar(core, user)

	res, err := registrar.ResolveAccountByKey()
	require.NoError(t, err, "Unexpected error resolving account by key")

	assert.Equal(t, "valid", res.Body.Status, "Unexpected account status")
}
