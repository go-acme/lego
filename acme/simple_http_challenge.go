package acme

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// OnSimpleHTTPStart hook will get called BEFORE SimpleHTTP starts to listen on a port.
var OnSimpleHTTPStart func(string)

// OnSimpleHTTPEnd hook will get called AFTER SimpleHTTP determined the status of the domain.
var OnSimpleHTTPEnd func(bool)

type simpleHTTPChallenge struct {
	jws     *jws
	optPort string
}

func (s *simpleHTTPChallenge) Solve(chlng challenge, domain string) error {

	logf("Trying to solve SimpleHTTP")

	// Generate random string for the path. The acme server will
	// access this path on the server in order to validate the request
	listener, err := s.startHTTPSServer(domain, chlng.Token)
	if err != nil {
		return fmt.Errorf("Could not start HTTPS server for challenge -> %v", err)
	}
	defer listener.Close()

	// Tell the server about the generated random path
	jsonBytes, err := json.Marshal(challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token})
	if err != nil {
		return errors.New("Failed to marshal network message...")
	}

	resp, err := s.jws.post(chlng.URI, jsonBytes)
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
			if OnSimpleHTTPEnd != nil {
				OnSimpleHTTPEnd(true)
			}
			logf("The server validated our request")
			break Loop
		case "pending":
			break
		case "invalid":
			if OnSimpleHTTPEnd != nil {
				OnSimpleHTTPEnd(false)
			}
			return errors.New("The server could not validate our request.")
		default:
			return errors.New("The server returned an unexpected state.")
		}

		time.Sleep(1 * time.Second)
		resp, err = http.Get(chlng.URI)
	}

	return nil
}

// Starts a temporary HTTPS server on port 443. As soon as the challenge passed validation,
// this server will get shut down. The certificate generated here is only held in memory.
func (s *simpleHTTPChallenge) startHTTPSServer(domain string, token string) (net.Listener, error) {

	// Generate a new RSA key and a self-signed certificate.
	tempPrivKey, err := generatePrivateKey(rsakey, 2048)
	rsaPrivKey := tempPrivKey.(*rsa.PrivateKey)
	if err != nil {
		return nil, err
	}
	tempCertPEM, err := generatePemCert(rsaPrivKey, domain)
	if err != nil {
		return nil, err
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivKey)})
	tempKeyPair, err := tls.X509KeyPair(
		tempCertPEM,
		pemBytes)
	if err != nil {
		return nil, err
	}

	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{tempKeyPair}

	path := "/.well-known/acme-challenge/" + token
	if OnSimpleHTTPStart != nil {
		OnSimpleHTTPStart(path)
	}

	// Allow for CLI override
	port := ":443"
	if s.optPort != "" {
		port = ":" + s.optPort
	}

	tlsListener, err := tls.Listen("tcp", domain+port, tlsConf)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(challenge{Type: "simpleHttp", Token: token, TLS: true})
	if err != nil {
		return nil, errors.New("startHTTPSServer: Failed to marshal network message")
	}
	signed, err := s.jws.signContent(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("startHTTPSServer: Failed to sign message. %s", err)
	}
	signedCompact := signed.FullSerialize()
	if err != nil {
		return nil, errors.New("startHTTPSServer: Failed to serialize message")
	}

	// The handler validates the HOST header and request type.
	// For validation it then writes the token the server returned with the challenge
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, domain) && r.Method == "GET" {
			w.Header().Add("Content-Type", "application/jose+json")
			w.Write([]byte(signedCompact))
			logf("Served JWS payload...")
		} else {
			logf("Received request for domain %s with method %s", r.Host, r.Method)
			w.Write([]byte("TEST"))
		}
	})

	go http.Serve(tlsListener, nil)

	return tlsListener, nil
}
