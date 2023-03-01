//go:build root
// +build root

//  this tests need to run as root to function

package nfqueue

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testpayload = []byte("this is the server behind")

func simpleHttp(port string) {
	sv := http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write(testpayload)
		}),
	}

	sv.ListenAndServe()
}

// this test our firewall rule doesn't hinder other webeserver traffic
func TestNotHinder(t *testing.T) {
	port := "31234"
	prv, _ := NewHttpDpiProvider("31234")
	go simpleHttp(port)
	prv.Present("labmdawork", "sampletoken", "keyauth")
	defer prv.CleanUp("labmdawork", "sampletoken", "keyauth")
	resp, err := http.Get("http://127.0.0.1:31234/hello")
	if err != nil {
		panic(err)
	}
	respBody, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, testpayload, respBody)

}

func TestFirewallSet(t *testing.T) {
	err := setFirewallRule(true, "12345")
	assert.Nil(t, err)
	defer setFirewallRule(false, "12345")
}

func TestChallengeinner(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)

	providerServer, _ := NewHttpDpiProvider("23457")
	go simpleHttp("23457")
	time.Sleep(50 * time.Microsecond)
	_, err := http.Get("http://127.0.0.1:23457/hello")
	assert.Nil(t, err)
	validate := func(_ *api.Core, _ string, chlng acme.Challenge) error {
		uri := "http://localhost" + ":23457" + http01.ChallengePath(chlng.Token)

		resp, err := http.DefaultClient.Get(uri)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if want := "text/plain"; resp.Header.Get("Content-Type") != want {
			t.Errorf("Get(%q) Content-Type: got %q, want %q", uri, resp.Header.Get("Content-Type"), want)
		}

		body, err := io.ReadAll(resp.Body)
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

	solver := http01.NewChallenge(core, validate, providerServer)

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
