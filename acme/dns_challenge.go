package acme

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type preCheckDNSFunc func(fqdn *Domain, value string) (bool, error)

var (
	preCheckDNS preCheckDNSFunc = checkDNSPropagation
	fqdnToZone                  = map[string]string{}
)

// dnsChallenge implements the dns-01 challenge according to ACME 7.5
type dnsChallenge struct {
	jws      *jws
	validate validateFunc
	provider ChallengeProvider
}

func (s *dnsChallenge) Solve(chlng challenge, domain string) error {
	logf("[INFO][%s] acme: Trying to solve DNS-01", domain)

	verifyDomain := NewDomain(domain)

	if s.provider == nil {
		return errors.New("No DNS Provider configured")
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, s.jws.privKey)
	if err != nil {
		return err
	}

	err = s.provider.Present(verifyDomain, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("Error presenting token %s", err)
	}
	defer func() {
		err := s.provider.CleanUp(verifyDomain, chlng.Token, keyAuth)
		if err != nil {
			log.Printf("Error cleaning up %s %v ", verifyDomain.Domain, err)
		}
	}()

	fqdn, value, _ := verifyDomain.GetDNS01Record(keyAuth)

	txtRecord := NewDomain(fqdn)

	logf("[INFO][%s] Checking DNS record propagation...", domain)

	var timeout, interval time.Duration
	switch provider := s.provider.(type) {
	case ChallengeProviderTimeout:
		timeout, interval = provider.Timeout()
	default:
		timeout, interval = 60*time.Second, 2*time.Second
	}

	err = WaitFor(timeout, interval, func() (bool, error) {
		return preCheckDNS(txtRecord, value)
	})
	if err != nil {
		return err
	}

	return s.validate(s.jws, domain, chlng.URI, challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

// checkDNSPropagation checks if the expected TXT record has been propagated to all authoritative nameservers.
func checkDNSPropagation(fqdn *Domain, value string) (bool, error) {
	return fqdn.CheckDNSPropagation(value)
}

// waitFor polls the given function 'f', once every 'interval' seconds, up to 'timeout' seconds.
func waitFor(timeout, interval int, f func() (bool, error)) error {
	var lastErr string
	timeup := time.After(time.Duration(timeout) * time.Second)
	for {
		select {
			case <-timeup:
				return fmt.Errorf("Time limit exceeded. Last error: %s", lastErr)
			default:
		}
		
		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			lastErr = err.Error()
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
