package http01

import (
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/emca/le/api"
	"github.com/xenolf/lego/platform/tester"
)

func TestChallenge(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
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

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
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
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	privKey, err := rsa.GenerateKey(rand.Reader, 128)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL, "", privKey)
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

// FIXME remove?
// stubValidate is like validate, except it does nothing.
func stubValidate(_ *api.Core, _, _ string, _ le.Challenge) error {
	return nil
}
