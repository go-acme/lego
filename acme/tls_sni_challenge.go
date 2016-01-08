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
	jws      *jws
	validate validateFunc
	iface    string
	port     string
}

func (t *tlsSNIChallenge) Solve(chlng challenge, domain string) error {
	// FIXME: https://github.com/ietf-wg-acme/acme/pull/22
	// Currently we implement this challenge to track boulder, not the current spec!

	logf("[INFO][%s] acme: Trying to solve TLS-SNI-01", domain)

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &t.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	cert, err := t.generateCertificate(keyAuth)
	if err != nil {
		return err
	}

	// Allow for CLI port override
	port := "443"
	if t.port != "" {
		port = t.port
	}

	iface := ""
	if t.iface != "" {
		iface = t.iface
	}

	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{cert}

	listener, err := tls.Listen("tcp", net.JoinHostPort(iface, port), tlsConf)
	if err != nil {
		return fmt.Errorf("Could not start HTTPS server for challenge -> %v", err)
	}
	defer listener.Close()

	go http.Serve(listener, nil)

	return t.validate(t.jws, domain, chlng.URI, challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
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
