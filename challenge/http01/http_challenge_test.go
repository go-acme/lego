package http01

import (
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/acme/api"
	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/platform/tester"
)

func TestChallenge(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	providerServer := &ProviderServer{port: "23457"}

	validate := func(_ *api.Core, _ string, chlng acme.Challenge) error {
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

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	solver := NewChallenge(core, validate, providerServer)

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "localhost:23457",
		},
		Challenges: []acme.Challenge{
			{Type: challenge.HTTP01.String(), Token: "http1"},
		},
	}

	err = solver.Solve(authz)
	require.NoError(t, err)
}

func TestChallengeInvalidPort(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	privateKey, err := rsa.GenerateKey(rand.Reader, 128)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	validate := func(_ *api.Core, _ string, _ acme.Challenge) error { return nil }

	solver := NewChallenge(core, validate, &ProviderServer{port: "123456"})

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "localhost:123456",
		},
		Challenges: []acme.Challenge{
			{Type: challenge.HTTP01.String(), Token: "http2"},
		},
	}

	err = solver.Solve(authz)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port")
	assert.Contains(t, err.Error(), "123456")
}
