package api

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderService_New(t *testing.T) {
	mux, apiURL := tester.SetupFakeAPI(t)

	// small value keeps test fast
	privateKey, errK := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, errK, "Could not generate test key")

	mux.HandleFunc("/newOrder", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body, err := readSignedBody(r, privateKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		order := acme.Order{}
		err = json.Unmarshal(body, &order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = tester.WriteJSONResponse(w, acme.Order{
			Status:      acme.StatusValid,
			Identifiers: order.Identifiers,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	core, err := New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	order, err := core.Orders.New([]string{"example.com"})
	require.NoError(t, err)

	expected := acme.ExtendedOrder{
		Order: acme.Order{
			Status:      "valid",
			Identifiers: []acme.Identifier{{Type: "dns", Value: "example.com"}},
		},
	}
	assert.Equal(t, expected, order)
}

func readSignedBody(r *http.Request, privateKey *rsa.PrivateKey) ([]byte, error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	jws, err := jose.ParseSigned(string(reqBody))
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
