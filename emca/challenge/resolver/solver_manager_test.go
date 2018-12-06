package resolver

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/challenge/http01"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/emca/le/api"
	"github.com/xenolf/lego/platform/tester"
	"gopkg.in/square/go-jose.v2"
)

func TestSolverManager_SetHTTP01Address(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", key)
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

	privKey, _ := rsa.GenerateKey(rand.Reader, 512)

	// validateNoBody reads the http.Request POST body, parses the JWS and validates it to read the body.
	// If there is an error doing this,
	// or if the JWS body is not the empty JSON payload "{}" or a POST-as-GET payload "" an error is returned.
	// We use this to verify challenge POSTs to the ts below do not send a JWS body.
	validateNoBody := func(r *http.Request) error {
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		jws, err := jose.ParseSigned(string(reqBody))
		if err != nil {
			return err
		}

		body, err := jws.Verify(&jose.JSONWebKey{
			Key:       privKey.Public(),
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

	mux.HandleFunc("/nonce", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Add("Replay-Nonce", "12345")
		w.Header().Add("Retry-After", "0")
	})

	mux.HandleFunc("/chlg", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if err := validateNoBody(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		st := statuses[0]
		statuses = statuses[1:]

		chlg := &le.Challenge{Type: "http-01", Status: st, URL: "http://example.com/", Token: "token"}
		if st == le.StatusInvalid {
			chlg.Error = &le.ProblemDetails{}
		}

		err := writeJSONResponse(w, chlg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
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
			statuses: []string{le.StatusValid},
		},
		{
			name:     "POST-invalid",
			statuses: []string{le.StatusInvalid},
			want:     "error",
		},
		{
			name:     "POST-pending-unexpected",
			statuses: []string{le.StatusPending, "weird"},
			want:     "unexpected",
		},
		{
			name:     "POST-pending-valid",
			statuses: []string{le.StatusPending, le.StatusValid},
		},
		{
			name:     "POST-pending-invalid",
			statuses: []string{le.StatusPending, le.StatusInvalid},
			want:     "error",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			statuses = test.statuses

			err := validate(core, "example.com", apiURL+"/chlg", le.Challenge{Type: "http-01", Token: "token"})
			if test.want == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.want)
			}
		})
	}
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
