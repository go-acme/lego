package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type preCheckDNSFunc func(domain, fqdn string) bool

var preCheckDNS preCheckDNSFunc = checkDNS

var preCheckDNSFallbackCount = 5

// DNS01Record returns a DNS record which will fulfill the `dns-01` challenge
func DNS01Record(domain, keyAuth string) (fqdn string, value string, ttl int) {
	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// base64URL encoding without padding
	keyAuthSha := base64.URLEncoding.EncodeToString(keyAuthShaBytes[:sha256.Size])
	value = strings.TrimRight(keyAuthSha, "=")
	ttl = 120
	fqdn = fmt.Sprintf("_acme-challenge.%s.", domain)
	return
}

// dnsChallenge implements the dns-01 challenge according to ACME 7.5
type dnsChallenge struct {
	jws      *jws
	validate validateFunc
	provider ChallengeProvider
}

func (s *dnsChallenge) Solve(chlng challenge, domain string) error {
	logf("[INFO][%s] acme: Trying to solve DNS-01", domain)

	if s.provider == nil {
		return errors.New("No DNS Provider configured")
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &s.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	err = s.provider.Present(domain, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("Error presenting token %s", err)
	}
	defer func() {
		err := s.provider.CleanUp(domain, chlng.Token, keyAuth)
		if err != nil {
			log.Printf("Error cleaning up %s %v ", domain, err)
		}
	}()

	fqdn, _, _ := DNS01Record(domain, keyAuth)

	preCheckDNS(domain, fqdn)

	return s.validate(s.jws, domain, chlng.URI, challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

func checkDNS(domain, fqdn string) bool {
	// check if the expected DNS entry was created. If not wait for some time and try again.
	m := new(dns.Msg)
	m.SetQuestion(domain+".", dns.TypeSOA)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, "google-public-dns-a.google.com:53")
	if err != nil {
		return false
	}

	var authorativeNS string
	for _, answ := range in.Answer {
		soa := answ.(*dns.SOA)
		authorativeNS = soa.Ns
	}

	fallbackCnt := 0
	for fallbackCnt < preCheckDNSFallbackCount {
		m.SetQuestion(fqdn, dns.TypeTXT)
		in, _, err = c.Exchange(m, authorativeNS+":53")
		if err != nil {
			return false
		}

		if len(in.Answer) > 0 {
			return true
		}

		fallbackCnt++
		if fallbackCnt >= preCheckDNSFallbackCount {
			return false
		}

		time.Sleep(time.Second * time.Duration(fallbackCnt))
	}

	return false
}

// toFqdn converts the name into a fqdn appending a trailing dot.
func toFqdn(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	}
	return name + "."
}

// unFqdn converts the fqdn into a name removing the trailing dot.
func unFqdn(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}
