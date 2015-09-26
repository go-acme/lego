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
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"
)

type simpleHTTPChallenge struct {
	jws     *jws
	optPort string
}

// SimpleHTTPS checks for DNS, public IP and port bindings
func (s *simpleHTTPChallenge) CanSolve(domain string) bool {

	// determine public ip
	resp, err := http.Get("https://icanhazip.com/")
	if err != nil {
		logger().Printf("Could not get public IP -> %v", err)
		return false
	}

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger().Printf("Could not get public IP -> %v", err)
		return false
	}
	ipStr := string(ip)
	ipStr = strings.Replace(ipStr, "\n", "", -1)

	// resolve domain we should solve for
	resolvedIPs, err := net.LookupHost(domain)
	if err != nil {
		logger().Printf("Could not lookup DNS A record for %s", domain)
		return false
	}

	// if the resolve does not resolve to our public ip, we can't solve.
	for _, resolvedIP := range resolvedIPs {
		if resolvedIP == ipStr {
			return true
		}
	}

	logger().Printf("SimpleHTTP: Domain %s does not resolve to the public ip of this server. Determined IP: %s Resolved IP: %s", domain, ipStr, resolvedIPs[0])
	return false
}

func (s *simpleHTTPChallenge) Solve(chlng challenge, domain string) error {

	logger().Print("Trying to solve SimpleHTTP")

	// Generate random string for the path. The acme server will
	// access this path on the server in order to validate the request
	listener, err := s.startHTTPSServer(domain, chlng.Token)
	if err != nil {
		return fmt.Errorf("Could not start HTTPS server for challenge -> %v", err)
	}

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
			listener.Close()
			logger().Print("The server could not validate our request.")
			return errors.New("The server could not validate our request.")
		default:
			listener.Close()
			logger().Print("The server returned an unexpected state.")
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
		return nil, err
	}

	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{tempKeyPair}

	path := "/.well-known/acme-challenge/" + token

	// Allow for CLI override
	port := ":443"
	if s.optPort != "" {
		port = ":" + s.optPort
	}

	tlsListener, err := tls.Listen("tcp", port, tlsConf)
	if err != nil {
		return nil, fmt.Errorf("Could not start HTTP listener! -> %v", err)
	}

	jsonBytes, err := json.Marshal(challenge{Type: "simpleHttp", Token: token, Tls: true})
	if err != nil {
		return nil, errors.New("startHTTPSServer: Failed to marshal network message...")
	}
	signed, err := s.jws.signContent(jsonBytes)
	if err != nil {
		return nil, errors.New("startHTTPSServer: Failed to sign message...")
	}
	signedCompact := signed.FullSerialize()
	if err != nil {
		return nil, errors.New("startHTTPSServer: Failed to serialize message...")
	}

	// The handler validates the HOST header and request type.
	// For validation it then writes the token the server returned with the challenge
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, domain) && r.Method == "GET" {
			w.Header().Add("Content-Type", "application/jose+json")
			w.Write([]byte(signedCompact))
			logger().Print("Served JWS payload...")
		} else {
			logger().Printf("Received request for domain %s with method %s", r.Host, r.Method)
			w.Write([]byte("TEST"))
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
