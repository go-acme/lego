package acme

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func TestTLSSNIChallenge(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	j := &jws{privKey: privKey.(*rsa.PrivateKey)}
	clientChallenge := challenge{Type: "tls-sni-01", Token: "tlssni1"}
	mockValidate := func(_ *jws, _, _ string, chlng challenge) error {
		conn, err := tls.Dial("tcp", "localhost:23457", &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Errorf("Expected to connect to challenge server without an error. %s", err.Error())
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

		zBytes := sha256.Sum256([]byte(chlng.KeyAuthorization))
		z := hex.EncodeToString(zBytes[:sha256.Size])
		domain := fmt.Sprintf("%s.%s.acme.invalid", z[:32], z[32:])

		if remoteCert.DNSNames[0] != domain {
			t.Errorf("Expected the challenge certificate DNSName to match %s but was %s", domain, remoteCert.DNSNames[0])
		}

		return nil
	}
	solver := &tlsSNIChallenge{jws: j, validate: mockValidate, port: "23457"}

	if err := solver.Solve(clientChallenge, "localhost:23457"); err != nil {
		t.Errorf("Solve error: got %v, want nil", err)
	}
}

func TestTLSSNIChallengeInvalidPort(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 128)
	j := &jws{privKey: privKey.(*rsa.PrivateKey)}
	clientChallenge := challenge{Type: "tls-sni-01", Token: "tlssni2"}
	solver := &tlsSNIChallenge{jws: j, validate: stubValidate, port: "123456"}

	if err := solver.Solve(clientChallenge, "localhost:123456"); err == nil {
		t.Errorf("Solve error: got %v, want error", err)
	} else if want := "invalid port 123456"; !strings.HasSuffix(err.Error(), want) {
		t.Errorf("Solve error: got %q, want suffix %q", err.Error(), want)
	}
}
