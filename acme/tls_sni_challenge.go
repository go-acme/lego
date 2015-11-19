package acme

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"net"
)

type tlsSNIChallenge struct {
	jws     *jws
	optPort string
	start   chan net.Listener
	end     chan error
}

func (t *tlsSNIChallenge) Solve(chlng challenge, domain string) error {

	logf("[INFO] acme: Trying to solve TLS-SNI-01")

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &t.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	zet := make([]string, chlng.Iterations)

	zetBytes := sha256.Sum256([]byte(keyAuth))
	zet[0] = hex.EncodeToString(zetBytes[:sha256.Size])
	for i := 1; i < chlng.Iterations; i++ {
		zetBytes = sha256.Sum256([]byte(zet[i-1]))
		zet[i] = hex.EncodeToString(zetBytes[:sha256.Size])
	}

	certificates, err := t.generateCertificates(zet)

	return nil
}

func (t *tlsSNIChallenge) generateCertificates(zet []string) ([]*x509.Certificate, error) {
	
	return nil, nil
}
