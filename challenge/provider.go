package challenge

import (
	"context"
	"time"

	"github.com/go-acme/lego/v5/acme"
)

// Provider enables implementing a custom challenge
// provider. Present presents the solution to a challenge available to
// be solved. CleanUp will be called by the challenge if Present ends
// in a non-error state.
type Provider interface {
	Present(ctx context.Context, domain, token, keyAuth string) error
	CleanUp(ctx context.Context, domain, token, keyAuth string) error
}

// ProviderTimeout allows for implementing a
// Provider where an unusually long timeout is required when
// waiting for an ACME challenge to be satisfied, such as when
// checking for DNS record propagation. If an implementor of a
// Provider provides a Timeout method, then the return values
// of the Timeout method will be used when appropriate by the acme
// package. The interval value is the time between checks.
//
// The default values used for timeout and interval are 60 seconds and
// 2 seconds respectively. These are used when no Timeout method is
// defined for the Provider.
type ProviderTimeout interface {
	Provider
	Timeout() (timeout, interval time.Duration)
}

// PersistentProvider enables implementing a custom challenge provider of DNS-PERSISTENT-01.
// IMPORTANT: this interface is experimental and may change without notice.
type PersistentProvider interface {
	Persist(ctx context.Context, authz acme.Authorization, issuerDomainName, accountURI string, persistUntil time.Time) error
	Timeout() (timeout, interval time.Duration)
}
