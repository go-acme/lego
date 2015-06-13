package acme

import (
	"crypto/rand"
	"crypto/rsa"
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
	caURL, optPort := "https://foobar", "1234"
	client := NewClient(caURL, user, keyBits, optPort)

	if client.jws == nil {
		t.Fatalf("Expected client.jws to not be nil")
	}
	if expected, actual := key, client.jws.privKey; actual != expected {
		t.Errorf("Expected jws.privKey to be %p but was %p", expected, actual)
	}

	if client.regURL != caURL {
		t.Errorf("Expected regURL to be '%s' but was '%s'", caURL, client.regURL)
	}
	if client.keyBits != keyBits {
		t.Errorf("Expected keyBits to be %d but was %d", keyBits, client.keyBits)
	}

	if expected, actual := 1, len(client.Solvers); actual != expected {
		t.Fatal("Expected %d solver(s), got %d", expected, actual)
	}

	simphttp, ok := client.Solvers["simpleHttps"].(*simpleHTTPChallenge)
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
