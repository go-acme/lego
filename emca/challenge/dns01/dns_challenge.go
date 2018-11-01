package dns01

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/xenolf/lego/emca/api"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/wait"
)

// FIXME refactor
type validateFunc func(core *api.Core, domain, uri string, chlng le.Challenge) error

const (
	// DefaultPropagationTimeout default propagation timeout
	DefaultPropagationTimeout = 60 * time.Second

	// DefaultPollingInterval default polling interval
	DefaultPollingInterval = 2 * time.Second

	// DefaultTTL default TTL
	DefaultTTL = 120
)

// Challenge implements the dns-01 challenge according to ACME 7.5
type Challenge struct {
	core     *api.Core
	validate validateFunc
	provider challenge.Provider
}

func NewChallenge(core *api.Core, validate validateFunc, provider challenge.Provider) *Challenge {
	return &Challenge{
		core:     core,
		validate: validate,
		provider: provider,
	}
}

// PreSolve just submits the txt record to the dns provider. It does not validate record propagation, or
// do anything at all with the acme server.
func (s *Challenge) PreSolve(chlng le.Challenge, domain string) error {
	log.Infof("[%s] acme: Preparing to solve DNS-01", domain)

	if s.provider == nil {
		return errors.New("no DNS Provider configured")
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := s.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = s.provider.Present(domain, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("error presenting token: %s", err)
	}

	return nil
}

func (s *Challenge) Solve(chlng le.Challenge, domain string) error {
	log.Infof("[%s] acme: Trying to solve DNS-01", domain)

	// Generate the Key Authorization for the challenge
	keyAuth, err := s.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	fqdn, value, _ := GetRecord(domain, keyAuth)

	log.Infof("[%s] Checking DNS record propagation using %+v", domain, RecursiveNameservers)

	var timeout, interval time.Duration
	switch provider := s.provider.(type) {
	case challenge.ProviderTimeout:
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

	return s.validate(s.core, domain, chlng.URL, le.Challenge{Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

// CleanUp cleans the challenge
func (s *Challenge) CleanUp(chlng le.Challenge, domain string) error {
	keyAuth, err := s.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}
	return s.provider.CleanUp(domain, chlng.Token, keyAuth)
}

// GetRecord returns a DNS record which will fulfill the `dns-01` challenge
func GetRecord(domain, keyAuth string) (fqdn string, value string, ttl int) {
	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// base64URL encoding without padding
	value = base64.RawURLEncoding.EncodeToString(keyAuthShaBytes[:sha256.Size])
	ttl = DefaultTTL
	fqdn = fmt.Sprintf("_acme-challenge.%s.", domain)
	return
}
