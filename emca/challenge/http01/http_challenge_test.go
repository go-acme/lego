package http01

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/api"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/le"
)

func TestChallenge(t *testing.T) {
	serverURL, tearDown := mockAPIEndpoint()
	defer tearDown()

	providerServer := &ProviderServer{port: "23457"}

	mockValidate := func(_ *api.Core, _, _ string, chlng le.Challenge) error {
		uri := "http://localhost" + providerServer.GetAddress() + ChallengePath(chlng.Token)

		resp, err := http.DefaultClient.Get(uri)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if want := "text/plain"; resp.Header.Get("Content-Type") != want {
			t.Errorf("Get(%q) Content-Type: got %q, want %q", uri, resp.Header.Get("Content-Type"), want)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyStr := string(body)

		if bodyStr != chlng.KeyAuthorization {
			t.Errorf("Get(%q) Body: got %q, want %q", uri, bodyStr, chlng.KeyAuthorization)
		}

		return nil
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", serverURL, "", privKey)
	require.NoError(t, err)

	solver := &Challenge{
		core:     core,
		validate: mockValidate,
		provider: providerServer,
	}

	clientChallenge := le.Challenge{Type: string(challenge.HTTP01), Token: "http1"}

	err = solver.Solve(clientChallenge, "localhost:23457")
	require.NoError(t, err)
}

func TestChallengeInvalidPort(t *testing.T) {
	serverURL, tearDown := mockAPIEndpoint()
	defer tearDown()

	privKey, err := rsa.GenerateKey(rand.Reader, 128)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", serverURL, "", privKey)
	require.NoError(t, err)

	solver := &Challenge{
		core:     core,
		validate: stubValidate,
		provider: &ProviderServer{port: "123456"},
	}

	clientChallenge := le.Challenge{Type: string(challenge.HTTP01), Token: "http2"}

	err = solver.Solve(clientChallenge, "localhost:123456")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port")
	assert.Contains(t, err.Error(), "123456")
}

func mockAPIEndpoint() (string, func()) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := writeJSONResponse(w, le.Directory{
			NewNonceURL:   ts.URL + "/nonce",
			NewAccountURL: ts.URL + "/account",
			NewOrderURL:   ts.URL + "/newOrder",
			RevokeCertURL: ts.URL + "/revokeCert",
			KeyChangeURL:  ts.URL + "/keyChange",
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return ts.URL, ts.Close
}

// FIXME remove?
// stubValidate is like validate, except it does nothing.
func stubValidate(_ *api.Core, _, _ string, _ le.Challenge) error {
	return nil
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
