package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	key, err := generatePrivateKey(32)
	if err != nil {
		t.Error("Error generating private key:", err)
	}
	if key == nil {
		t.Error("Expected key to not be nil, but it was")
	}
}

func TestGenerateCSR(t *testing.T) {
	key, err := generatePrivateKey(512)
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	csr, err := generateCsr(key, "fizz.buzz")
	if err != nil {
		t.Error("Error generating CSR:", err)
	}
	if csr == nil || len(csr) == 0 {
		t.Error("Expected CSR with data, but it was nil or length 0")
	}
}

func TestPEMEncode(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 32)
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	data := pemEncode(key)

	if data == nil {
		t.Fatal("Expected result to not be nil, but it was")
	}
	if len(data) != 127 {
		t.Errorf("Expected PEM encoding to be length 127, but it was %d", len(data))
	}
}
