package acme

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
)

// tlsSNIChallengeServer implements ChallengeProvider for `TLS-SNI-01` challenge
type tlsSNIChallengeServer struct {
	iface    string
	port     string
	done     chan bool
	listener net.Listener
}

// Present makes the keyAuth available as a cert
func (s *tlsSNIChallengeServer) Present(domain, token, keyAuth string) error {
	if s.port == "" {
		s.port = "443"
	}

	cert, err := TLSSNI01ChallengeCert(keyAuth)
	if err != nil {
		return err
	}

	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{cert}

	s.listener, err = tls.Listen("tcp", net.JoinHostPort(s.iface, s.port), tlsConf)
	if err != nil {
		return fmt.Errorf("Could not start HTTPS server for challenge -> %v", err)
	}

	s.done = make(chan bool)
	go func() {
		http.Serve(s.listener, nil)
		s.done <- true
	}()
	return nil
}

func (s *tlsSNIChallengeServer) CleanUp(domain, token, keyAuth string) error {
	if s.listener == nil {
		return nil
	}
	s.listener.Close()
	<-s.done
	return nil
}
