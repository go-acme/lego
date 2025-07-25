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
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	var statuses []string

	privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)

	apiURL := tester.MockACMEServer().
		Route("POST /chlg",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if err := validateNoBody(privateKey, req); err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				rw.Header().Set("Link",
					fmt.Sprintf(`<http://%s/my-authz>; rel="up"`, req.Context().Value(http.LocalAddrContextKey)))

				st := statuses[0]
				statuses = statuses[1:]

				chlg := &acme.Challenge{Type: "http-01", Status: st, URL: "http://example.com/", Token: "token"}

				servermock.JSONEncode(chlg).ServeHTTP(rw, req)
			})).
		Route("POST /my-authz",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				st := statuses[0]
				statuses = statuses[1:]

				authorization := acme.Authorization{
					Status:     st,
					Challenges: []acme.Challenge{},
				}

				if st == acme.StatusInvalid {
					chlg := acme.Challenge{
						Status: acme.StatusInvalid,
					}
					authorization.Challenges = append(authorization.Challenges, chlg)
				}

				servermock.JSONEncode(authorization).ServeHTTP(rw, req)
			})).
		Build(t)

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
			want:     "the server returned an unexpected challenge status: weird",
		},
		{
			name:     "POST-valid",
			statuses: []string{acme.StatusValid},
		},
		{
			name:     "POST-invalid",
			statuses: []string{acme.StatusInvalid},
			want:     "invalid challenge:",
		},
		{
			name:     "POST-pending-unexpected",
			statuses: []string{acme.StatusPending, "weird"},
			want:     "the server returned an unexpected authorization status: weird",
		},
		{
			name:     "POST-pending-valid",
			statuses: []string{acme.StatusPending, acme.StatusValid},
		},
		{
			name:     "POST-pending-invalid",
			statuses: []string{acme.StatusPending, acme.StatusInvalid},
			want:     "invalid authorization",
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

func Test_checkChallengeStatus(t *testing.T) {
	testCases := []struct {
		desc       string
		challenge  acme.Challenge
		requireErr require.ErrorAssertionFunc
		expected   bool
	}{
		{
			desc:       "status valid",
			challenge:  acme.Challenge{Status: acme.StatusValid},
			requireErr: require.NoError,
			expected:   true,
		},
		{
			desc:       "status invalid",
			challenge:  acme.Challenge{Status: acme.StatusInvalid},
			requireErr: require.Error,
			expected:   false,
		},
		{
			desc:       "status invalid with error",
			challenge:  acme.Challenge{Status: acme.StatusInvalid, Error: &acme.ProblemDetails{}},
			requireErr: require.Error,
			expected:   false,
		},
		{
			desc:       "status pending",
			challenge:  acme.Challenge{Status: acme.StatusPending},
			requireErr: require.NoError,
			expected:   false,
		},
		{
			desc:       "status processing",
			challenge:  acme.Challenge{Status: acme.StatusProcessing},
			requireErr: require.NoError,
			expected:   false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			status, err := checkChallengeStatus(acme.ExtendedChallenge{Challenge: test.challenge})
			test.requireErr(t, err)

			assert.Equal(t, test.expected, status)
		})
	}
}

func Test_checkAuthorizationStatus(t *testing.T) {
	testCases := []struct {
		desc          string
		authorization acme.Authorization
		requireErr    require.ErrorAssertionFunc
		expected      bool
	}{
		{
			desc:          "status valid",
			authorization: acme.Authorization{Status: acme.StatusValid},
			requireErr:    require.NoError,
			expected:      true,
		},
		{
			desc:          "status invalid",
			authorization: acme.Authorization{Status: acme.StatusInvalid},
			requireErr:    require.Error,
			expected:      false,
		},
		{
			desc:          "status invalid with error",
			authorization: acme.Authorization{Status: acme.StatusInvalid, Challenges: []acme.Challenge{{Error: &acme.ProblemDetails{}}}},
			requireErr:    require.Error,
			expected:      false,
		},
		{
			desc:          "status pending",
			authorization: acme.Authorization{Status: acme.StatusPending},
			requireErr:    require.NoError,
			expected:      false,
		},
		{
			desc:          "status processing",
			authorization: acme.Authorization{Status: acme.StatusProcessing},
			requireErr:    require.NoError,
			expected:      false,
		},
		{
			desc:          "status deactivated",
			authorization: acme.Authorization{Status: acme.StatusDeactivated},
			requireErr:    require.Error,
			expected:      false,
		},
		{
			desc:          "status expired",
			authorization: acme.Authorization{Status: acme.StatusExpired},
			requireErr:    require.Error,
			expected:      false,
		},
		{
			desc:          "status revoked",
			authorization: acme.Authorization{Status: acme.StatusRevoked},
			requireErr:    require.Error,
			expected:      false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			status, err := checkAuthorizationStatus(test.authorization)
			test.requireErr(t, err)

			assert.Equal(t, test.expected, status)
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

	sigAlgs := []jose.SignatureAlgorithm{jose.RS256}
	jws, err := jose.ParseSigned(string(reqBody), sigAlgs)
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
