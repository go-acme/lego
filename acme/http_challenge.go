package acme

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type httpChallenge struct {
	jws      *jws
	validate validateFunc
	optPort  string
}

func (s *httpChallenge) Solve(chlng challenge, domain string) error {

	logf("[INFO][%s] acme: Trying to solve HTTP-01", domain)

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &s.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	// Allow for CLI port override
	port := ":80"
	if s.optPort != "" {
		port = ":" + s.optPort
	}

	listener, err := net.Listen("tcp", domain+port)
	if err != nil {
		// if the domain:port bind failed, fall back to :port bind and try that instead.
		listener, err = net.Listen("tcp", port)
		if err != nil {
			return fmt.Errorf("Could not start HTTP server for challenge -> %v", err)
		}
	}
	defer listener.Close()

	path := "/.well-known/acme-challenge/" + chlng.Token

	// The handler validates the HOST header and request type.
	// For validation it then writes the token the server returned with the challenge
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, domain) && r.Method == "GET" {
			w.Header().Add("Content-Type", "text/plain")
			w.Write([]byte(keyAuth))
			logf("[INFO] Served key authentication")
		} else {
			logf("[INFO] Received request for domain %s with method %s", r.Host, r.Method)
			w.Write([]byte("TEST"))
		}
	})

	go http.Serve(listener, mux)

	return s.validate(s.jws, domain, chlng.URI, challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}
