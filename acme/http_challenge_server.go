package acme

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// httpChallengeServer implements ChallengeProvider for `http-01` challenge
type httpChallengeServer struct {
	iface    string
	port     string
	done     chan bool
	listener net.Listener
}

// Present makes the token available at `HTTP01ChallengePath(token)`
func (s *httpChallengeServer) Present(domain, token, keyAuth string) error {
	if s.port == "" {
		s.port = "80"
	}

	var err error
	s.listener, err = net.Listen("tcp", net.JoinHostPort(s.iface, s.port))
	if err != nil {
		return fmt.Errorf("Could not start HTTP server for challenge -> %v", err)
	}

	s.done = make(chan bool)
	go s.serve(domain, token, keyAuth)
	return nil
}

func (s *httpChallengeServer) CleanUp(domain, token, keyAuth string) error {
	if s.listener == nil {
		return nil
	}
	s.listener.Close()
	<-s.done
	return nil
}

func (s *httpChallengeServer) serve(domain, token, keyAuth string) {
	path := HTTP01ChallengePath(token)

	// The handler validates the HOST header and request type.
	// For validation it then writes the token the server returned with the challenge
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, domain) && r.Method == "GET" {
			w.Header().Add("Content-Type", "text/plain")
			w.Write([]byte(keyAuth))
			logf("[INFO][%s] Served key authentication", domain)
		} else {
			logf("[INFO] Received request for domain %s with method %s", r.Host, r.Method)
			w.Write([]byte("TEST"))
		}
	})

	http.Serve(s.listener, mux)
	s.done <- true
}
