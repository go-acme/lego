package dns01

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

// checkNameserversPropagation queries each of the recursive nameservers for the expected TXT record.
func (c *Client) checkNameserversPropagation(ctx context.Context, fqdn, value string, addPort bool) (bool, error) {
	return c.checkNameserversPropagationCustom(ctx, fqdn, value, c.recursiveNameservers, addPort)
}

// checkNameserversPropagationCustom queries each of the given nameservers for the expected TXT record.
func (c *Client) checkNameserversPropagationCustom(ctx context.Context, fqdn, value string, nameservers []string, addPort bool) (bool, error) {
	for _, ns := range nameservers {
		if addPort {
			ns = net.JoinHostPort(ns, c.authoritativeNSPort)
		}

		r, err := c.sendQueryCustom(ctx, fqdn, dns.TypeTXT, []string{ns}, false)
		if err != nil {
			return false, err
		}

		if r.Rcode != dns.RcodeSuccess {
			return false, fmt.Errorf("NS %s returned %s for %s", ns, dns.RcodeToString[r.Rcode], fqdn)
		}

		var records []string

		var found bool

		for _, rr := range r.Answer {
			if txt, ok := rr.(*dns.TXT); ok {
				record := strings.Join(txt.Txt, "")

				records = append(records, record)
				if record == value {
					found = true
					break
				}
			}
		}

		if !found {
			return false, fmt.Errorf("NS %s did not return the expected TXT record [fqdn: %s, value: %s]: %s", ns, fqdn, value, strings.Join(records, ", "))
		}
	}

	return true, nil
}

// lookupAuthoritativeNameservers returns the authoritative nameservers for the given fqdn.
func (c *Client) lookupAuthoritativeNameservers(ctx context.Context, fqdn string) ([]string, error) {
	var authoritativeNss []string

	zone, err := c.FindZoneByFqdn(ctx, fqdn)
	if err != nil {
		return nil, fmt.Errorf("could not find zone: %w", err)
	}

	r, err := c.sendQuery(ctx, zone, dns.TypeNS, true)
	if err != nil {
		return nil, fmt.Errorf("NS call failed: %w", err)
	}

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
