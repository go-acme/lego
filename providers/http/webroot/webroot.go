// Package webroot implements a HTTP provider for solving the HTTP-01 challenge using web server's root path.
package webroot

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/xenolf/lego/acme"
	"path/filepath"
)

// HTTPProvider implements ChallengeProvider for `http-01` challenge
type HTTPProvider struct {
	path string
}

// NewHTTPProvider returns a HTTPProvider instance with a configured webroot path
func NewHTTPProvider(path string) (*HTTPProvider, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("Webroot path does not exist")
	}

	c := &HTTPProvider{
		path: path,
	}

	return c, nil
}

// Present makes the token available at `HTTP01ChallengePath(token)` by creating a file in the given webroot path
func (w *HTTPProvider) Present(domain, token, keyAuth string) error {
	var err error

	challengeFilePath := path.Join(w.path, acme.HTTP01ChallengePath(token))
	err = os.MkdirAll(path.Dir(challengeFilePath), 0777)
	if err != nil {
		return fmt.Errorf("Could not create required directories in webroot for HTTP challenge -> %v", err)
	}

	err = ioutil.WriteFile(challengeFilePath, []byte(keyAuth), 0777)
	if err != nil {
		return fmt.Errorf("Could not write file in webroot for HTTP challenge -> %v", err)
	}

	return nil
}

// CleanUp removes the file created for the challenge
func (w *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	var err error
	p := filepath.Join(w.path, acme.HTTP01ChallengePath(token))
	for len(p) > len(w.path) {
		err = os.Remove(p)
		if err != nil {
			return fmt.Errorf("Could not remove file in webroot after HTTP challenge -> %v", err)
		}
		p = filepath.Dir(p)
	}

	return nil
}
