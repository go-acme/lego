package dns01

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

// PreCheckFunc checks DNS propagation before notifying ACME that the DNS challenge is ready.
type PreCheckFunc func(fqdn, value string) (bool, error)

func AddPreCheck(preCheck PreCheckFunc) ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheckDNSFunc = preCheck
		return nil
	}
}

// checkDNSPropagation checks if the expected TXT record has been propagated to all authoritative nameservers.
func checkDNSPropagation(fqdn, value string) (bool, error) {
	// Initial attempt to resolve at the recursive NS
	r, err := dnsQuery(fqdn, dns.TypeTXT, recursiveNameservers, true)
	if err != nil {
		return false, err
	}

	if r.Rcode == dns.RcodeSuccess {
		// If we see a CNAME here then use the alias
		for _, rr := range r.Answer {
			if cn, ok := rr.(*dns.CNAME); ok {
				if cn.Hdr.Name == fqdn {
					fqdn = cn.Target
					break
				}
			}
		}
	}

	authoritativeNss, err := lookupNameservers(fqdn)
	if err != nil {
		return false, err
	}

	return checkAuthoritativeNss(fqdn, value, authoritativeNss)
}

// checkAuthoritativeNss queries each of the given nameservers for the expected TXT record.
func checkAuthoritativeNss(fqdn, value string, nameservers []string) (bool, error) {
	for _, ns := range nameservers {
		r, err := dnsQuery(fqdn, dns.TypeTXT, []string{net.JoinHostPort(ns, "53")}, false)
		if err != nil {
			return false, err
		}

		if r.Rcode != dns.RcodeSuccess {
			return false, fmt.Errorf("NS %s returned %s for %s", ns, dns.RcodeToString[r.Rcode], fqdn)
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
			return false, fmt.Errorf("NS %s did not return the expected TXT record [fqdn: %s]", ns, fqdn)
		}
	}

	return true, nil
}
