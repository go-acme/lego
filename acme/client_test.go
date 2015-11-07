package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal("Could not generate test key:", err)
	}
	user := mockUser{
		email:      "test@test.com",
		regres:     new(RegistrationResource),
		privatekey: key,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(directory{NewAuthzURL: "http://test", NewCertURL: "http://test", NewRegURL: "http://test", RevokeCertURL: "http://test"})
		w.Write(data)
	}))

	caURL, optPort := ts.URL, "1234"
	client, err := NewClient(caURL, user, keyBits, optPort)
	if err != nil {
		t.Fatalf("Could not create client: %v", err)
	}

	if client.jws == nil {
		t.Fatalf("Expected client.jws to not be nil")
	}
	if expected, actual := key, client.jws.privKey; actual != expected {
		t.Errorf("Expected jws.privKey to be %p but was %p", expected, actual)
	}

	if client.keyBits != keyBits {
		t.Errorf("Expected keyBits to be %d but was %d", keyBits, client.keyBits)
	}

	if expected, actual := 1, len(client.solvers); actual != expected {
		t.Fatalf("Expected %d solver(s), got %d", expected, actual)
	}

	simphttp, ok := client.solvers["simpleHttp"].(*simpleHTTPChallenge)
	if !ok {
		t.Fatal("Expected simpleHttps solver to be simpleHTTPChallenge type")
	}
	if simphttp.jws != client.jws {
		t.Error("Expected simpleHTTPChallenge to have same jws as client")
	}
	if simphttp.optPort != optPort {
		t.Errorf("Expected simpleHTTPChallenge to have optPort %s but was %s", optPort, simphttp.optPort)
	}
}

type mockUser struct {
	email      string
	regres     *RegistrationResource
	privatekey *rsa.PrivateKey
}

func (u mockUser) GetEmail() string                       { return u.email }
func (u mockUser) GetRegistration() *RegistrationResource { return u.regres }
func (u mockUser) GetPrivateKey() *rsa.PrivateKey         { return u.privatekey }
