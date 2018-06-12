package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/asn1"
	"strings"
	"testing"
)

func TestTLSALPNChallenge(t *testing.T) {
	domain := "localhost:23457"
	privKey, _ := rsa.GenerateKey(rand.Reader, 512)
	j := &jws{privKey: privKey}
	clientChallenge := challenge{Type: string(TLSALPN01), Token: "tlsalpn1"}
	mockValidate := func(_ *jws, _, _ string, chlng challenge) error {
		conn, err := tls.Dial("tcp", domain, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Errorf("Expected to connect to challenge server without an error. %v", err)
		}

		// Expect the server to only return one certificate
		connState := conn.ConnectionState()
		if count := len(connState.PeerCertificates); count != 1 {
			t.Errorf("Expected the challenge server to return exactly one certificate but got %d", count)
		}

		remoteCert := connState.PeerCertificates[0]
		if count := len(remoteCert.DNSNames); count != 1 {
			t.Errorf("Expected the challenge certificate to have exactly one DNSNames entry but had %d", count)
		}

		if remoteCert.DNSNames[0] != domain {
			t.Errorf("Expected the challenge certificate DNSName to match %s but was %s", domain, remoteCert.DNSNames[0])
		}

		if len(remoteCert.Extensions) == 0 {
			t.Error("Expected the challenge certificate to contain extensions, it contained nothing")
		}

		idx := -1
		for i, ext := range remoteCert.Extensions {
			if idPeAcmeIdentifierV1.Equal(ext.Id) {
				idx = i
				break
			}
		}

		if idx == -1 {
			t.Fatal("Expected the challenge certificate to contain an extension with the id-pe-acmeIdentifier id, it did not")
		}

		ext := remoteCert.Extensions[idx]

		if !ext.Critical {
			t.Error("Expected the challenge certificate id-pe-acmeIdentifier extension to be marked as critical, it was not")
		}

		zBytes := sha256.Sum256([]byte(chlng.KeyAuthorization))
		value, err := asn1.Marshal(zBytes[:sha256.Size])
		if err != nil {
			t.Fatalf("Expected marshaling of the keyAuth to return no error, but was %v", err)
		}
		if subtle.ConstantTimeCompare(value[:], ext.Value) != 1 {
			t.Errorf("Expected the challenge certificate id-pe-acmeIdentifier extension to contain the SHA-256 digest of the keyAuth, %v, but was %v", zBytes[:], ext.Value)
		}

		return nil
	}
	solver := &tlsALPNChallenge{jws: j, validate: mockValidate, provider: &TLSALPNProviderServer{port: "23457"}}
	if err := solver.Solve(clientChallenge, domain); err != nil {
		t.Errorf("Solve error: got %v, want nil", err)
	}
}

func TestTLSALPNChallengeInvalidPort(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 128)
	j := &jws{privKey: privKey}
	clientChallenge := challenge{Type: string(TLSALPN01), Token: "tlsalpn1"}
	solver := &tlsALPNChallenge{jws: j, validate: stubValidate, provider: &TLSALPNProviderServer{port: "123456"}}

	if err := solver.Solve(clientChallenge, "localhost:123456"); err == nil {
		t.Errorf("Solve error: got %v, want error", err)
	} else if want, want18 := "invalid port 123456", "123456: invalid port"; !strings.HasSuffix(err.Error(), want) && !strings.HasSuffix(err.Error(), want18) {
		t.Errorf("Solve error: got %q, want suffix %q", err.Error(), want)
	}
}
