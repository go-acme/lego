package tlsalpn01

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/log"
)

const (
	// ACMETLS1Protocol is the ALPN Protocol ID for the ACME-TLS/1 Protocol.
	ACMETLS1Protocol = "acme-tls/1"

	// defaultTLSPort is the port that the ProviderServer will default to
	// when no other port is provided.
	defaultTLSPort = "443"
)

var _ challenge.Provider = (*ProviderServer)(nil)

type Options struct {
	Network      string
	NetworkStack challenge.NetworkStack
	Host         string
	Port         string
}

// ProviderServer implements ChallengeProvider for `TLS-ALPN-01` challenge.
// It may be instantiated without using the NewProviderServer
// if you want only to use the default values.
type ProviderServer struct {
	network string
	address string

	listener net.Listener
}

// NewProviderServerWithOptions creates a new ProviderServer.
func NewProviderServerWithOptions(opts Options) *ProviderServer {
	if opts.Port == "" {
		// Fallback to port 443 if the port was not provided.
		opts.Port = defaultTLSPort
	}

	if opts.Network == "" {
		opts.Network = "tcp"
	}

	return &ProviderServer{
		network: opts.NetworkStack.Network(opts.Network),
		address: net.JoinHostPort(opts.Host, opts.Port),
	}
}

// NewProviderServer creates a new ProviderServer on the selected interface and port.
// Setting host and / or port to an empty string will make the server fall back to
// the "any" interface and port 443 respectively.
func NewProviderServer(host, port string) *ProviderServer {
	return NewProviderServerWithOptions(Options{Host: host, Port: port})
}

// Present generates a certificate with an SHA-256 digest of the keyAuth provided
// as the acmeValidation-v1 extension value to conform to the ACME-TLS-ALPN spec.
func (s *ProviderServer) Present(ctx context.Context, domain, token, keyAuth string) error {
	// Generate the challenge certificate using the provided keyAuth and domain.
	cert, err := ChallengeCert(domain, keyAuth)
	if err != nil {
		return err
	}

	// Place the generated certificate with the extension into the TLS config
	// so that it can serve the correct details.
	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{*cert}

	// We must set that the `acme-tls/1` application level protocol is supported
	// so that the protocol negotiation can succeed. Reference:
	// https://www.rfc-editor.org/rfc/rfc8737.html#section-6.2
	tlsConf.NextProtos = []string{ACMETLS1Protocol}

	// Create the listener with the created tls.Config.
	s.listener, err = tls.Listen(s.network, s.GetAddress(), tlsConf)
	if err != nil {
		return fmt.Errorf("could not start HTTPS server for challenge: %w", err)
	}

	// Shut the server down when we're finished.
	go func() {
		err := http.Serve(s.listener, nil)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Warn("HTTP server serve.", log.ErrorAttr(err))
		}
	}()

	return nil
}

// CleanUp closes the HTTPS server.
func (s *ProviderServer) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	if s.listener == nil {
		return nil
	}

	// Server was created, close it.
	if err := s.listener.Close(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *ProviderServer) GetAddress() string {
	return s.address
}
