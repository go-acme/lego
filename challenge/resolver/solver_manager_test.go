package resolver

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/acme/api"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/challenge/http01"
	"github.com/xenolf/lego/platform/tester"
	"gopkg.in/square/go-jose.v2"
)

func TestSolverManager_SetHTTP01Address(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	solversManager := NewSolversManager(core)

	optPort := "1234"
	optHost := ""

	err = solversManager.SetHTTP01Address(net.JoinHostPort(optHost, optPort))
	require.NoError(t, err)

	require.IsType(t, &http01.Challenge{}, solversManager.solvers[challenge.HTTP01])
	httpSolver := solversManager.solvers[challenge.HTTP01].(*http01.Challenge)

	httpProviderServer := (*http01.ProviderServer)(unsafe.Pointer(reflect.ValueOf(httpSolver).Elem().FieldByName("provider").InterfaceData()[1]))
	assert.Equal(t, net.JoinHostPort(optHost, optPort), httpProviderServer.GetAddress())

	// test setting different host
	optHost = "127.0.0.1"
	err = solversManager.SetHTTP01Address(net.JoinHostPort(optHost, optPort))
	require.NoError(t, err)

	httpProviderServer = (*http01.ProviderServer)(unsafe.Pointer(reflect.ValueOf(httpSolver).Elem().FieldByName("provider").InterfaceData()[1]))
	assert.Equal(t, net.JoinHostPort(optHost, optPort), httpProviderServer.GetAddress())
}

func TestValidate(t *testing.T) {
	mux, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	var statuses []string

	privateKey, _ := rsa.GenerateKey(rand.Reader, 512)

	mux.HandleFunc("/chlg", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if err := validateNoBody(privateKey, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Link", "<"+apiURL+`/my-authz>; rel="up"`)

		st := statuses[0]
		statuses = statuses[1:]

		chlg := &acme.Challenge{Type: "http-01", Status: st, URL: "http://example.com/", Token: "token"}
		if st == acme.StatusInvalid {
			chlg.Error = &acme.ProblemDetails{}
		}

		err := tester.WriteJSONResponse(w, chlg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/my-authz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		st := statuses[0]
		statuses = statuses[1:]

		authorization := acme.Authorization{
			Status:     st,
			Challenges: []acme.Challenge{},
		}

		if st == acme.StatusInvalid {
			chlg := acme.Challenge{
				Status: acme.StatusInvalid,
				Error:  &acme.ProblemDetails{},
			}
			authorization.Challenges = append(authorization.Challenges, chlg)
		}

		err := tester.WriteJSONResponse(w, authorization)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		statuses []string
		want     string
	}{
		{
			name:     "POST-unexpected",
			statuses: []string{"weird"},
			want:     "unexpected",
		},
		{
			name:     "POST-valid",
			statuses: []string{acme.StatusValid},
		},
		{
			name:     "POST-invalid",
			statuses: []string{acme.StatusInvalid},
			want:     "error",
		},
		{
			name:     "POST-pending-unexpected",
			statuses: []string{acme.StatusPending, "weird"},
			want:     "unexpected",
		},
		{
			name:     "POST-pending-valid",
			statuses: []string{acme.StatusPending, acme.StatusValid},
		},
		{
			name:     "POST-pending-invalid",
			statuses: []string{acme.StatusPending, acme.StatusInvalid},
			want:     "error",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			statuses = test.statuses

			err := validate(core, "example.com", acme.Challenge{Type: "http-01", Token: "token", URL: apiURL + "/chlg"})
			if test.want == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.want)
			}
		})
	}
}

// validateNoBody reads the http.Request POST body, parses the JWS and validates it to read the body.
// If there is an error doing this,
// or if the JWS body is not the empty JSON payload "{}" or a POST-as-GET payload "" an error is returned.
// We use this to verify challenge POSTs to the ts below do not send a JWS body.
func validateNoBody(privateKey *rsa.PrivateKey, r *http.Request) error {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	jws, err := jose.ParseSigned(string(reqBody))
	if err != nil {
		return err
	}

	body, err := jws.Verify(&jose.JSONWebKey{
		Key:       privateKey.Public(),
		Algorithm: "RSA",
	})
	if err != nil {
		return err
	}

	if bodyStr := string(body); bodyStr != "{}" && bodyStr != "" {
		return fmt.Errorf(`expected JWS POST body "{}" or "", got %q`, bodyStr)
	}
	return nil
}
