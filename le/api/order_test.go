package api

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/platform/tester"
	"gopkg.in/square/go-jose.v2"
)

func TestOrderService_New(t *testing.T) {
	mux, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	// small value keeps test fast
	privKey, errK := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, errK, "Could not generate test key")

	mux.HandleFunc("/newOrder", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body, err := readSignedBody(r, privKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		order := le.Order{}
		err = json.Unmarshal(body, &order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = tester.WriteJSONResponse(w, le.Order{
			Status:      le.StatusValid,
			Identifiers: order.Identifiers,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	core, err := New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privKey)
	require.NoError(t, err)

	order, err := core.Orders.New([]string{"example.com"})
	require.NoError(t, err)

	expected := le.ExtendedOrder{
		Order: le.Order{
			Status:      "valid",
			Identifiers: []le.Identifier{{Type: "dns", Value: "example.com"}},
		},
	}
	assert.Equal(t, expected, order)
}

func readSignedBody(r *http.Request, privKey *rsa.PrivateKey) ([]byte, error) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	jws, err := jose.ParseSigned(string(reqBody))
	if err != nil {
		return nil, err
	}

	body, err := jws.Verify(&jose.JSONWebKey{
		Key:       privKey.Public(),
		Algorithm: "RSA",
	})
	if err != nil {
		return nil, err
	}

	return body, nil
}
