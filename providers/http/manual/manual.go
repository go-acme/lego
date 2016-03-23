// Package manual implements an HTTP provider for solving the HTTP-01
// challenge by displaying instructions necessary to manually satisfy
// the challenge.
package manual

import (
	"bufio"
	"log"
	"os"

	"github.com/xenolf/lego/acme"
)

// logf writes a log entry. It uses acme.Logger if not
// nil, otherwise it uses the default log.Logger.
func logf(format string, args ...interface{}) {
	if acme.Logger != nil {
		acme.Logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// HTTPProvider is an implementation of the acme.ChallengeProvider
// interface.
type HTTPProvider struct{}

// NewHTTPProvider returns an HTTPProvider instance.
func NewHTTPProvider() (*HTTPProvider, error) {
	return &HTTPProvider{}, nil
}

// Present prints instructions for manually configuring a web server
// for the http-01 challenge.
func (*HTTPProvider) Present(domain, token, keyAuth string) error {
	logf("[INFO] acme: Please configure your web server so that an HTTP GET request to")
	logf("[INFO] acme: http://%s%s", domain, acme.HTTP01ChallengePath(token))
	logf("[INFO] acme: returns the following string in the response body")
	logf("[INFO] acme: %s", keyAuth)
	logf("[INFO] acme: Press 'Enter' when you are done")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	return nil
}

// CleanUp prints instructions for manually configuring a web server
// once the http-01 challenge is completed.
func (*HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	logf("[INFO] acme: You can now remove the following URL from your server configuration")
	logf("[INFO] acme: http://%s%s", domain, acme.HTTP01ChallengePath(token))
	return nil
}
