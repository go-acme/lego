package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type preCheckDNSFunc func(domain, fqdn, value string) error

var preCheckDNS preCheckDNSFunc = checkDnsPropagation

var recursionMaxDepth = 10

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

	fqdn, value, _ := DNS01Record(domain, keyAuth)

	logf("[INFO][%s] Checking DNS record propagation...", domain)

	if err = preCheckDNS(domain, fqdn, value); err != nil {
		return err
	}

	return s.validate(s.jws, domain, chlng.URI, challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

// checkDnsPropagation checks if the expected DNS entry has been propagated to
// all authoritative nameservers. If not it waits and retries for some time.
func checkDnsPropagation(domain, fqdn, value string) error {
	authoritativeNss, err := lookupNameservers(toFqdn(domain))
	if err != nil {
		return err
	}

	if err = waitFor(30, 2, func() (bool, error) {
		return checkAuthoritativeNss(fqdn, value, authoritativeNss)
	}); err != nil {
		return err
	}

	return nil
}

// checkAuthoritativeNss checks whether a TXT record with fqdn and value exists on every given nameserver.
func checkAuthoritativeNss(fqdn, value string, nameservers []string) (bool, error) {
	for _, ns := range nameservers {
		r, err := dnsQuery(fqdn, dns.TypeTXT, net.JoinHostPort(ns, "53"))
		if err != nil {
			return false, err
		}

		if r.Rcode != dns.RcodeSuccess {
			return false, fmt.Errorf("%s returned RCode %s", ns, dns.RcodeToString[r.Rcode])
		}

		var found bool
		for _, rr := range r.Answer {
			if txt, ok := rr.(*dns.TXT); ok {
				if strings.Join(txt.Txt, "") == value {
					found = true
					break
				}
			}
		}
		if !found {
			return false, fmt.Errorf("%s did not return the expected TXT record", ns)
		}
	}

	return true, nil
}

// dnsQuery directly queries the given authoritative nameserver.
func dnsQuery(fqdn string, rtype uint16, nameserver string) (in *dns.Msg, err error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, rtype)
	m.SetEdns0(4096, false)
	m.RecursionDesired = false

	in, err = dns.Exchange(m, nameserver)
	if err == dns.ErrTruncated {
		tcp := &dns.Client{Net: "tcp"}
		in, _, err = tcp.Exchange(m, nameserver)
	}

	return
}

// lookupNameservers returns the authoritative nameservers for the given domain name.
func lookupNameservers(fqdn string) ([]string, error) {
	var referralNameservers []string

	// We start recursion at the gTLD origin
	// so we don't have to manage root hints
	labels := dns.SplitDomainName(fqdn)
	tld := labels[len(labels)-1]
	nss, err := net.LookupNS(tld)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve TLD %s %v", tld, err)
	}

	for _, ns := range nss {
		referralNameservers = append(referralNameservers, ns.Host)
	}

	// Follow the referrals until we hit the authoritative NS
	for depth := 0; depth < recursionMaxDepth; depth++ {
		var r *dns.Msg
		var err error

		for _, ns := range referralNameservers {
			r, err = dnsQuery(fqdn, dns.TypeNS, net.JoinHostPort(ns, "53"))
			if err != nil {
				continue
			}

			if r.Rcode == dns.RcodeSuccess {
				break
			}

			if r.Rcode == dns.RcodeNameError  {
				return nil, fmt.Errorf("Could not resolve NXDOMAIN %s", fqdn)
			}
		}

		if r == nil {
			break
		}

		if r.Authoritative {
			// We got an authoritative reply, which means that the
			// last referral holds the authoritative nameservers.
			return referralNameservers, nil
		}

		referralNameservers = nil

		for _, rr := range r.Ns {
			if ns, ok := rr.(*dns.NS); ok {
				referralNameservers = append(referralNameservers, strings.ToLower(ns.Ns))
			}
		}

		// No referrals to follow
		if len(referralNameservers) == 0 {
			break
		}
	}

	return nil, fmt.Errorf("Could not determine nameservers for %s", fqdn)
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
