package dnspersist01

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// defaultNameserverPort used by authoritative NS.
var defaultNameserverPort = "53"

// RecordMatcher returns true when the expected record is present.
type RecordMatcher func(records []TXTRecord) bool

// PreCheckFunc checks DNS propagation before notifying ACME that the challenge is ready.
type PreCheckFunc func(fqdn string, matcher RecordMatcher) (bool, error)

// WrapPreCheckFunc wraps a PreCheckFunc in order to do extra operations before or after
// the main check, put it in a loop, etc.
type WrapPreCheckFunc func(domain, fqdn string, matcher RecordMatcher, check PreCheckFunc) (bool, error)

// WrapPreCheck Allow to define checks before notifying ACME that the challenge is ready.
func WrapPreCheck(wrap WrapPreCheckFunc) ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.checkFunc = wrap
		return nil
	}
}

// DisableAuthoritativeNssPropagationRequirement disables authoritative nameserver checks.
func DisableAuthoritativeNssPropagationRequirement() ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.requireAuthoritativeNssPropagation = false
		return nil
	}
}

// RecursiveNSsPropagationRequirement requires checks against recursive nameservers.
func RecursiveNSsPropagationRequirement() ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.requireRecursiveNssPropagation = true
		return nil
	}
}

// AddRecursiveNameservers overrides recursive nameservers used for propagation checks.
func AddRecursiveNameservers(nameservers []string) ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.recursiveNameservers = ParseNameservers(nameservers)
		return nil
	}
}

// PropagationWait sleeps for the specified duration, optionally skipping checks.
func PropagationWait(wait time.Duration, skipCheck bool) ChallengeOption {
	return WrapPreCheck(func(domain, fqdn string, matcher RecordMatcher, check PreCheckFunc) (bool, error) {
		time.Sleep(wait)

		if skipCheck {
			return true, nil
		}

		return check(fqdn, matcher)
	})
}

type preCheck struct {
	// checks DNS propagation before notifying ACME that the DNS challenge is ready.
	checkFunc WrapPreCheckFunc

	// require the TXT record to be propagated to all authoritative name servers
	requireAuthoritativeNssPropagation bool

	// require the TXT record to be propagated to all recursive name servers
	requireRecursiveNssPropagation bool
}

func newPreCheck() preCheck {
	return preCheck{
		requireAuthoritativeNssPropagation: true,
	}
}

func (p preCheck) call(domain, fqdn string, matcher RecordMatcher, check PreCheckFunc) (bool, error) {
	if p.checkFunc == nil {
		return check(fqdn, matcher)
	}

	return p.checkFunc(domain, fqdn, matcher, check)
}

func (c *Challenge) checkDNSPropagation(fqdn string, matcher RecordMatcher) (bool, error) {
	nameservers := c.getRecursiveNameservers()

	// Initial attempt to resolve at the recursive NS (require to get CNAME)
	result, err := c.resolver.lookupTXT(fqdn, nameservers, true)
	if err != nil {
		return false, fmt.Errorf("initial recursive nameserver: %w", err)
	}

	effectiveFQDN := dns.Fqdn(fqdn)
	if len(result.CNAMEChain) > 0 {
		effectiveFQDN = result.CNAMEChain[len(result.CNAMEChain)-1]
	}

	if c.preCheck.requireRecursiveNssPropagation {
		_, err = c.checkNameserversPropagation(effectiveFQDN, nameservers, false, true, matcher)
		if err != nil {
			return false, fmt.Errorf("recursive nameservers: %w", err)
		}
	}

	if !c.preCheck.requireAuthoritativeNssPropagation {
		return true, nil
	}

	authoritativeNss, err := lookupNameservers(effectiveFQDN, nameservers, c.resolver.Timeout)
	if err != nil {
		return false, err
	}

	found, err := c.checkNameserversPropagation(effectiveFQDN, authoritativeNss, true, false, matcher)
	if err != nil {
		return found, fmt.Errorf("authoritative nameservers: %w", err)
	}

	return found, nil
}

func (c *Challenge) checkNameserversPropagation(fqdn string, nameservers []string, addPort, recursive bool, matcher RecordMatcher) (bool, error) {
	for _, ns := range nameservers {
		if addPort {
			ns = net.JoinHostPort(ns, defaultNameserverPort)
		}

		result, err := c.resolver.lookupTXT(fqdn, []string{ns}, recursive)
		if err != nil {
			return false, err
		}

		if !matcher(result.Records) {
			return false, fmt.Errorf("NS %s did not return a matching TXT record [fqdn: %s]: %s", ns, fqdn, strings.Join(txtValues(result.Records), " ,"))
		}
	}

	return true, nil
}

func txtValues(records []TXTRecord) []string {
	values := make([]string, 0, len(records))
	for _, record := range records {
		values = append(values, record.Value)
	}

	return values
}

// lookupNameservers returns the authoritative nameservers for the given fqdn.
func lookupNameservers(fqdn string, nameservers []string, timeout time.Duration) ([]string, error) {
	zone, err := findZoneByFqdn(fqdn, nameservers, timeout)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	r, err := dnsQueryWithTimeout(zone, dns.TypeNS, nameservers, true, timeout)
	if err != nil {
		return nil, fmt.Errorf("NS call failed: %w", err)
	}

	var authoritativeNss []string

	for _, rr := range r.Answer {
		if ns, ok := rr.(*dns.NS); ok {
			authoritativeNss = append(authoritativeNss, strings.ToLower(ns.Ns))
		}
	}

	if len(authoritativeNss) > 0 {
		return authoritativeNss, nil
	}

	return nil, fmt.Errorf("[zone=%s] could not determine authoritative nameservers", zone)
}

func findZoneByFqdn(fqdn string, nameservers []string, timeout time.Duration) (string, error) {
	var (
		err error
		r   *dns.Msg
	)

	for _, domain := range domainsSeq(fqdn) {
		r, err = dnsQueryWithTimeout(domain, dns.TypeSOA, nameservers, true, timeout)
		if err != nil {
			continue
		}

		if r == nil {
			continue
		}

		switch r.Rcode {
		case dns.RcodeSuccess:
			// Check if we got a SOA RR in the answer section
			if len(r.Answer) == 0 {
				continue
			}

			// CNAME records cannot/should not exist at the root of a zone.
			// So we skip a domain when a CNAME is found.
			if dnsMsgContainsCNAME(r) {
				continue
			}

			for _, ans := range r.Answer {
				if soa, ok := ans.(*dns.SOA); ok {
					return soa.Hdr.Name, nil
				}
			}
		case dns.RcodeNameError:
			// NXDOMAIN
		default:
			// Any response code other than NOERROR and NXDOMAIN is treated as error
			return "", &DNSError{Message: fmt.Sprintf("unexpected response for '%s'", domain), MsgOut: r}
		}
	}

	return "", &DNSError{Message: fmt.Sprintf("could not find the start of authority for '%s'", dns.Fqdn(fqdn)), MsgOut: r, Err: err}
}

func dnsMsgContainsCNAME(msg *dns.Msg) bool {
	for _, ans := range msg.Answer {
		if _, ok := ans.(*dns.CNAME); ok {
			return true
		}
	}

	return false
}

func domainsSeq(fqdn string) []string {
	fqdn = dns.Fqdn(fqdn)
	if fqdn == "" {
		return nil
	}

	var domains []string
	for _, index := range dns.Split(fqdn) {
		domains = append(domains, fqdn[index:])
	}

	return domains
}
