package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"
)

type simpleHTTPChallenge struct {
	jws     *jws
	optPort string
}

func (s *simpleHTTPChallenge) CanSolve() bool {
	return true
}

func (s *simpleHTTPChallenge) Solve(chlng challenge, domain string) error {

	logger().Print("Trying to solve SimpleHTTPS")

	responseToken := getRandomString(15)
	listener, err := s.startHTTPSServer(domain, chlng.Token, responseToken)
	if err != nil {
		return fmt.Errorf("Could not start HTTPS server for challenge -> %v", err)
	}

	jsonBytes, err := json.Marshal(challenge{Type: chlng.Type, Path: responseToken})
	if err != nil {
		return errors.New("Failed to marshal network message...")
	}

	resp, err := s.jws.post(chlng.URI, jsonBytes)
	if err != nil {
		return fmt.Errorf("Failed to post JWS message. -> %v", err)
	}

	var challengeResponse challenge
loop:
	for {
		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&challengeResponse)

		switch challengeResponse.Status {
		case "valid":
			logger().Print("The server validated our request")
			listener.Close()
			break loop
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

// Starts a temporary HTTPS server on port 443. As soon as the challenge passed validation,
// this server will get shut down. The certificate generated here is only held in memory.
func (s *simpleHTTPChallenge) startHTTPSServer(domain string, token string, responseToken string) (net.Listener, error) {
	tempPrivKey, err := generatePrivateKey(2048)
	if err != nil {
		return nil, err
	}
	tempCertPEM, err := generateCert(tempPrivKey, domain)
	if err != nil {
		return nil, err
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(tempPrivKey)})
	tempKeyPair, err := tls.X509KeyPair(
		tempCertPEM,
		pemBytes)
	if err != nil {
		logger().Print("error here!")
		return nil, err
	}

	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{tempKeyPair}

	path := "/.well-known/acme-challenge/" + responseToken

	port := ":443"
	if s.optPort != "" {
		port = ":" + s.optPort
	}
	tlsListener, err := tls.Listen("tcp", port, tlsConf)
	if err != nil {
		logger().Fatalf("Could not start HTTP listener! -> %v", err)
	}

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Host == domain && r.Method == "GET" {
			w.Write([]byte(token))
		}
	})

	go http.Serve(tlsListener, nil)

	return tlsListener, nil
}

func getRandomString(length int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func generateCert(privKey *rsa.PrivateKey, domain string) ([]byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "ACME Challenge TEMP",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365),

		KeyUsage:              x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}), nil
}
