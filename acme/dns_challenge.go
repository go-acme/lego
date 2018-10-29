package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/wait"
)

const (
	// DefaultPropagationTimeout default propagation timeout
	DefaultPropagationTimeout = 60 * time.Second

	// DefaultPollingInterval default polling interval
	DefaultPollingInterval = 2 * time.Second

	// DefaultTTL default TTL
	DefaultTTL = 120
)

// dnsChallenge implements the dns-01 challenge according to ACME 7.5
type dnsChallenge struct {
	jws      *jws
	validate validateFunc
	provider ChallengeProvider
}

// PreSolve just submits the txt record to the dns provider. It does not validate record propagation, or
// do anything at all with the acme server.
func (s *dnsChallenge) PreSolve(chlng challenge, domain string) error {
	log.Infof("[%s] acme: Preparing to solve DNS-01", domain)

	if s.provider == nil {
		return errors.New("no DNS Provider configured")
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := s.jws.getKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = s.provider.Present(domain, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("error presenting token: %s", err)
	}

	return nil
}

func (s *dnsChallenge) Solve(chlng challenge, domain string) error {
	log.Infof("[%s] acme: Trying to solve DNS-01", domain)

	// Generate the Key Authorization for the challenge
	keyAuth, err := s.jws.getKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	fqdn, value, _ := DNS01Record(domain, keyAuth)

	log.Infof("[%s] Checking DNS record propagation using %+v", domain, RecursiveNameservers)

	var timeout, interval time.Duration
	switch provider := s.provider.(type) {
	case ChallengeProviderTimeout:
		timeout, interval = provider.Timeout()
	default:
		timeout, interval = DefaultPropagationTimeout, DefaultPollingInterval
	}

	err = wait.For(timeout, interval, func() (bool, error) {
		return PreCheckDNS(fqdn, value)
	})
	if err != nil {
		return err
	}

	return s.validate(s.jws, domain, chlng.URL, challenge{Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

// CleanUp cleans the challenge
func (s *dnsChallenge) CleanUp(chlng challenge, domain string) error {
	keyAuth, err := s.jws.getKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}
	return s.provider.CleanUp(domain, chlng.Token, keyAuth)
}

// DNS01Record returns a DNS record which will fulfill the `dns-01` challenge
func DNS01Record(domain, keyAuth string) (fqdn string, value string, ttl int) {
	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// base64URL encoding without padding
	value = base64.RawURLEncoding.EncodeToString(keyAuthShaBytes[:sha256.Size])
	ttl = DefaultTTL
	fqdn = fmt.Sprintf("_acme-challenge.%s.", domain)
	return
}

// ToFqdn converts the name into a fqdn appending a trailing dot.
func ToFqdn(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	}
	return name + "."
}

// UnFqdn converts the fqdn into a name removing the trailing dot.
func UnFqdn(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}
