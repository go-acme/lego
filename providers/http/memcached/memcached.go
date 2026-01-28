// Package memcached implements an HTTP provider for solving the HTTP-01 challenge using memcached in combination with a webserver.
package memcached

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-acme/lego/v5/challenge/http01"
)

// HTTPProvider implements HTTPProvider for `http-01` challenge.
type HTTPProvider struct {
	hosts []string
}

// NewMemcachedProvider returns a HTTPProvider instance with a configured webroot path.
func NewMemcachedProvider(hosts []string) (*HTTPProvider, error) {
	if len(hosts) == 0 {
		return nil, errors.New("no memcached hosts provided")
	}

	return &HTTPProvider{hosts: hosts}, nil
}

// Present makes the token available at `HTTP01ChallengePath(token)` by creating a file in the given webroot path.
func (w *HTTPProvider) Present(_ context.Context, _, token, keyAuth string) error {
	var errs []error

	challengePath := path.Join("/", http01.ChallengePath(token))

	for _, host := range w.hosts {
		mc := memcache.New(host)

		// Only because this is slow on GitHub action.
		mc.Timeout = 1 * time.Second

		item := &memcache.Item{
			Key:        challengePath,
			Value:      []byte(keyAuth),
			Expiration: 60,
		}

		err := mc.Add(item)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) == len(w.hosts) {
		return fmt.Errorf("unable to store key in any of the memcached hosts: %w", errors.Join(errs...))
	}

	return nil
}

// CleanUp removes the file created for the challenge.
func (w *HTTPProvider) CleanUp(_ context.Context, _, _, _ string) error {
	// Memcached will clean up itself, that's what expiration is for.
	return nil
}
