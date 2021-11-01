package resolver

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"sort"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
)

func TestByType(t *testing.T) {
	challenges := []acme.Challenge{
		{Type: "dns-01"}, {Type: "tlsalpn-01"}, {Type: "http-01"},
	}

	sort.Sort(byType(challenges))

	expected := []acme.Challenge{
		{Type: "tlsalpn-01"}, {Type: "http-01"}, {Type: "dns-01"},
	}

	assert.Equal(t, expected, challenges)
}

func TestValidate(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

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
	reqBody, err := io.ReadAll(r.Body)
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
