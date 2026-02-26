package dns01

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

// checkRecursiveNameserversPropagation queries each of the recursive nameservers for the expected TXT record.
func (c *Client) checkRecursiveNameserversPropagation(ctx context.Context, fqdn, value string) (bool, error) {
	return c.checkNameserversPropagationCustom(ctx, fqdn, value, c.core.GetRecursiveNameservers(), false)
}

// checkRecursiveNameserversPropagation queries each of the authoritative nameservers for the expected TXT record.
func (c *Client) checkAuthoritativeNameserversPropagation(ctx context.Context, fqdn, value string) (bool, error) {
	authoritativeNss, err := c.core.LookupAuthoritativeNameservers(ctx, fqdn)
	if err != nil {
		return false, err
	}

	return c.checkNameserversPropagationCustom(ctx, fqdn, value, authoritativeNss, true)
}

// checkNameserversPropagationCustom queries each of the given nameservers for the expected TXT record.
func (c *Client) checkNameserversPropagationCustom(ctx context.Context, fqdn, value string, nameservers []string, addPort bool) (bool, error) {
	for _, ns := range nameservers {
		if addPort {
			ns = net.JoinHostPort(ns, c.authoritativeNSPort)
		}

		r, err := c.core.SendQueryCustom(ctx, fqdn, dns.TypeTXT, []string{ns}, false)
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
