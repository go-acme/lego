package acme

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
)

type tlsSNIChallenge struct {
	jws     *jws
	optPort string
	start   chan net.Listener
	end     chan error
}

func (t *tlsSNIChallenge) Solve(chlng challenge, domain string) error {
	// FIXME: https://github.com/ietf-wg-acme/acme/pull/22
	// Currently we implement this challenge to track boulder, not the current spec!

	logf("[INFO] acme: Trying to solve TLS-SNI-01")

	t.start = make(chan net.Listener)
	t.end = make(chan error)

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &t.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	certificate, err := t.generateCertificate(keyAuth)
	if err != nil {
		return err
	}

	go t.startSNITLSServer(certificate)
	var listener net.Listener
	select {
	case listener = <-t.start:
		break
	case err := <-t.end:
		return fmt.Errorf("Could not start HTTPS server for challenge -> %v", err)
	}

	// Make sure we properly close the HTTP server before we return
	defer func() {
		listener.Close()
		err = <-t.end
		close(t.start)
		close(t.end)
	}()

	return validate(t.jws, chlng.URI, challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

func (t *tlsSNIChallenge) generateCertificate(keyAuth string) (tls.Certificate, error) {

	zBytes := sha256.Sum256([]byte(keyAuth))
	z := hex.EncodeToString(zBytes[:sha256.Size])

	// generate a new RSA key for the certificates
	tempPrivKey, err := generatePrivateKey(rsakey, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	rsaPrivKey := tempPrivKey.(*rsa.PrivateKey)
	rsaPrivPEM := pemEncode(rsaPrivKey)

	domain := fmt.Sprintf("%s.%s.acme.invalid", z[:32], z[32:])
	tempCertPEM, err := generatePemCert(rsaPrivKey, domain)
	if err != nil {
		return tls.Certificate{}, err
	}

	certificate, err := tls.X509KeyPair(tempCertPEM, rsaPrivPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return certificate, nil
}

func (t *tlsSNIChallenge) startSNITLSServer(cert tls.Certificate) {

	// Allow for CLI port override
	port := ":443"
	if t.optPort != "" {
		port = ":" + t.optPort
	}

	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{cert}

	tlsListener, err := tls.Listen("tcp", port, tlsConf)
	if err != nil {
		t.end <- err
	}
	// Signal successfull start
	t.start <- tlsListener

	http.Serve(tlsListener, nil)

	t.end <- nil
}
