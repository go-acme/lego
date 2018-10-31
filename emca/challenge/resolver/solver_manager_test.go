package resolver

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/challenge/http01"
	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/internal/sender"
	"github.com/xenolf/lego/emca/le"
)

func TestSolverManager_SetHTTPAddress(t *testing.T) {
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

	user := mockUser{
		email:      "test@test.com",
		regres:     new(le.RegistrationResource),
		privatekey: key,
	}

	do := sender.NewDo(http.DefaultClient, "lego-test")
	solversManager := NewSolversManager(secure.NewJWS(do, user.GetPrivateKey(), ts.URL+"/nonce"))

	optPort := "1234"
	optHost := ""

	err = solversManager.SetHTTPAddress(net.JoinHostPort(optHost, optPort))
	require.NoError(t, err)

	require.IsType(t, &http01.Challenge{}, solversManager.solvers[challenge.HTTP01])
	httpSolver := solversManager.solvers[challenge.HTTP01].(*http01.Challenge)

	jws := (*secure.JWS)(unsafe.Pointer(reflect.ValueOf(httpSolver).Elem().FieldByName("jws").Pointer()))
	assert.Equal(t, solversManager.jws, jws, "Expected http-01 to have same jws as client")

	httpProviderServer := (*http01.ProviderServer)(unsafe.Pointer(reflect.ValueOf(httpSolver).Elem().FieldByName("provider").InterfaceData()[1]))
	assert.Equal(t, net.JoinHostPort(optHost, optPort), httpProviderServer.GetAddress())

	// test setting different host
	optHost = "127.0.0.1"
	err = solversManager.SetHTTPAddress(net.JoinHostPort(optHost, optPort))
	require.NoError(t, err)

	httpProviderServer = (*http01.ProviderServer)(unsafe.Pointer(reflect.ValueOf(httpSolver).Elem().FieldByName("provider").InterfaceData()[1]))
	assert.Equal(t, net.JoinHostPort(optHost, optPort), httpProviderServer.GetAddress())
}

type mockUser struct {
	email      string
	regres     *le.RegistrationResource
	privatekey *rsa.PrivateKey
}

func (u mockUser) GetEmail() string                          { return u.email }
func (u mockUser) GetRegistration() *le.RegistrationResource { return u.regres }
func (u mockUser) GetPrivateKey() crypto.PrivateKey          { return u.privatekey }
