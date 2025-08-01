package api

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderService_NewWithOptions(t *testing.T) {
	// small value keeps test fast
	privateKey, errK := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, errK, "Could not generate test key")

	apiURL, client := tester.MockACMEServer().
		Route("POST /newOrder",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				body, err := readSignedBody(req, privateKey)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				order := acme.Order{}
				err = json.Unmarshal(body, &order)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				servermock.JSONEncode(acme.Order{
					Status:         acme.StatusValid,
					Expires:        order.Expires,
					Identifiers:    order.Identifiers,
					Profile:        order.Profile,
					NotBefore:      order.NotBefore,
					NotAfter:       order.NotAfter,
					Error:          order.Error,
					Authorizations: order.Authorizations,
					Finalize:       order.Finalize,
					Certificate:    order.Certificate,
					Replaces:       order.Replaces,
				}).ServeHTTP(rw, req)
			})).
		BuildHTTPS(t)

	core, err := New(client, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	testCases := []struct {
		desc     string
		opts     *OrderOptions
		expected acme.ExtendedOrder
	}{
		{
			desc: "simple",
			expected: acme.ExtendedOrder{
				Order: acme.Order{
					Status:      "valid",
					Identifiers: []acme.Identifier{{Type: "dns", Value: "example.com"}},
				},
			},
		},
		{
			desc: "with options",
			opts: &OrderOptions{
				NotBefore: time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC),
				NotAfter:  time.Date(2023, 1, 2, 1, 0, 0, 0, time.UTC),
			},
			expected: acme.ExtendedOrder{
				Order: acme.Order{
					Status:      "valid",
					Identifiers: []acme.Identifier{{Type: "dns", Value: "example.com"}},
					NotBefore:   "2023-01-01T01:00:00Z",
					NotAfter:    "2023-01-02T01:00:00Z",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			order, err := core.Orders.NewWithOptions([]string{"example.com"}, test.opts)
			require.NoError(t, err)

			assert.Equal(t, test.expected, order)
		})
	}
}

func readSignedBody(r *http.Request, privateKey *rsa.PrivateKey) ([]byte, error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	sigAlgs := []jose.SignatureAlgorithm{jose.RS256}
	jws, err := jose.ParseSigned(string(reqBody), sigAlgs)
	if err != nil {
		return nil, err
	}

	body, err := jws.Verify(&jose.JSONWebKey{
		Key:       privateKey.Public(),
		Algorithm: "RSA",
	})
	if err != nil {
		return nil, err
	}

	return body, nil
}
