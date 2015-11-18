package acme

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type httpChallengeWebserver struct {
	optPort string
	start   chan net.Listener
	end     chan error
}

func (s *httpChallengeWebserver) PresentToken(domain, token, keyAuth string, checkSolvedFunc func() error) (err error) {
	s.start = make(chan net.Listener)
	s.end = make(chan error)

	go s.startHTTPServer(domain, token, keyAuth)
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

	err = checkSolvedFunc()
	return
}

func (s *httpChallengeWebserver) startHTTPServer(domain string, token string, keyAuth string) {

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
			logf("Served Key Authentication ...")
		} else {
			logf("Received request for domain %s with method %s", r.Host, r.Method)
			w.Write([]byte("TEST"))
		}
	})

	http.Serve(listener, nil)

	// Signal that the server was shut down
	s.end <- nil
}
