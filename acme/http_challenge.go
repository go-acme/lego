package acme

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type httpChallenge struct {
	jws     *jws
	optPort string
	webRoot string
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

	if s.webRoot == "" {
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
	} else {
		err := s.sendToken(chlng.Token, keyAuth)
		if err != nil {
			return err
		}

	}
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

func (s *httpChallenge) sendToken(token, keyAuth string) error {
	u, err := url.Parse("//" + s.webRoot)
	if err != nil {
		return fmt.Errorf("Could not parse the webRoot: %v", s.webRoot)
	}
	var auths []ssh.AuthMethod
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}
	username := os.Getenv("USER")
	if u.User != nil {
		p, b := u.User.Password()
		if b {
			auths = append(auths, ssh.Password(p))
		}
		username = u.User.Username()
	}

	config := ssh.ClientConfig{
		User: username,
		Auth: auths,
	}
	host := u.Host
	if !strings.ContainsRune(host, ':') {
		host = host + ":22"
	}
	conn, err := ssh.Dial("tcp", host, &config)
	if err != nil {
		return fmt.Errorf("unable to connect to [%s]: %v", host, err)
	}
	defer conn.Close()

	c, err := sftp.NewClient(conn, sftp.MaxPacket(1<<15))
	if err != nil {
		return fmt.Errorf("unable to start sftp subsytem: %v", err)
	}
	defer c.Close()
	bContent := []byte(keyAuth)
	w, err := c.Create(u.Path + "/.well-known/acme-challenge/" + token)
	if err != nil {
		return err
	}
	defer w.Close()

	n, err := w.Write(bContent)
	if err != nil {
		return err
	}
	if n != len(bContent) {
		return fmt.Errorf("copy: expected %v bytes, got %d", len(bContent), n)
	}
	return nil
}
