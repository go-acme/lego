package acme

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type httpChallenge struct {
	jws     *jws
	optPort string
	start   chan net.Listener
	end     chan error
}

func (s *httpChallenge) Solve(chlng challenge, domain string) error {

	logf("[INFO] acme: Trying to solve HTTP-01")

	s.start = make(chan net.Listener)
	s.end = make(chan error)

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &s.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	go s.startHTTPServer(domain, chlng.Token, keyAuth)
	var listener net.Listener
	select {
	case listener = <-s.start:
		break
	case err := <-s.end:
		return fmt.Errorf("Could not start HTTP server for challenge -> %v", err)
	}

	// Make sure we properly close the HTTP server before we return
	defer func() {
		listener.Close()
		err = <-s.end
		close(s.start)
		close(s.end)
	}()

	jsonBytes, err := json.Marshal(challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
	if err != nil {
		return errors.New("Failed to marshal network message...")
	}

	// Tell the server we handle HTTP-01
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

func (s *httpChallenge) startHTTPServer(domain string, token string, keyAuth string) {

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
			s.end <- err
		}
	}
	// Signal successfull start
	s.start <- listener

	path := "/.well-known/acme-challenge/" + token

	// The handler validates the HOST header and request type.
	// For validation it then writes the token the server returned with the challenge
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, domain) && r.Method == "GET" {
			w.Header().Add("Content-Type", "text/plain")
			w.Write([]byte(keyAuth))
			logf("[INFO] Served key authentication")
		} else {
			logf("[INFO] Received request for domain %s with method %s", r.Host, r.Method)
			w.Write([]byte("TEST"))
		}
	})

	http.Serve(listener, nil)

	// Signal that the server was shut down
	s.end <- nil
}
