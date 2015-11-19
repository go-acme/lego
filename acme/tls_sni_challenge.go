package acme

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
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

	jsonBytes, err := json.Marshal(challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
	if err != nil {
		return errors.New("Failed to marshal network message...")
	}

	// Tell the server we handle TLS-SNI-01
	resp, err := t.jws.post(chlng.URI, jsonBytes)
	if err != nil {
		return fmt.Errorf("Failed to post JWS message. -> %v", err)
	}

	// After the path is sent, the ACME server will access our server.
	// Repeatedly check the server for an updated status on our request.
	var challengeResponse challenge
Loop:
	for {
		if resp.StatusCode >= http.StatusBadRequest {
			return handleHTTPError(resp)
		}

		err = json.NewDecoder(resp.Body).Decode(&challengeResponse)
		resp.Body.Close()
		if err != nil {
			return err
		}

		switch challengeResponse.Status {
		case "valid":
			logf("The server validated our request")
			break Loop
		case "pending":
			break
		case "invalid":
			return errors.New("The server could not validate our request.")
		default:
			return errors.New("The server returned an unexpected state.")
		}

		time.Sleep(1 * time.Second)
		resp, err = http.Get(chlng.URI)
	}

	return nil
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
