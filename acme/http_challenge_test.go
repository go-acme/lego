package acme

import (
	"crypto/rsa"
	"io/ioutil"
	"strings"
	"testing"
)

func TestHTTPChallenge(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	j := &jws{privKey: privKey.(*rsa.PrivateKey)}
	clientChallenge := challenge{Type: "http-01", Token: "http1"}
	mockValidate := func(_ *jws, _, _ string, chlng challenge) error {
		uri := "http://localhost:23457/.well-known/acme-challenge/" + chlng.Token
		resp, err := httpGet(uri)
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
	solver := &httpChallenge{jws: j, validate: mockValidate, port: "23457"}

	if err := solver.Solve(clientChallenge, "localhost:23457"); err != nil {
		t.Errorf("Solve error: got %v, want nil", err)
	}
}

func TestHTTPChallengeInvalidPort(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 128)
	j := &jws{privKey: privKey.(*rsa.PrivateKey)}
	clientChallenge := challenge{Type: "http-01", Token: "http2"}
	solver := &httpChallenge{jws: j, validate: stubValidate, port: "123456"}

	if err := solver.Solve(clientChallenge, "localhost:123456"); err == nil {
		t.Errorf("Solve error: got %v, want error", err)
	} else if want := "invalid port 123456"; !strings.HasSuffix(err.Error(), want) {
		t.Errorf("Solve error: got %q, want suffix %q", err.Error(), want)
	}
}
