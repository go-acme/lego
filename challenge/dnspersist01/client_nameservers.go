package dnspersist01

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

const defaultResolvConf = "/etc/resolv.conf"

func (c *Client) checkNameserversPropagation(ctx context.Context, fqdn string, addPort, recursive bool, matcher RecordMatcher) (bool, error) {
	return c.checkNameserversPropagationCustom(ctx, fqdn, c.recursiveNameservers, addPort, recursive, matcher)
}

func (c *Client) checkNameserversPropagationCustom(ctx context.Context, fqdn string, nameservers []string, addPort, recursive bool, matcher RecordMatcher) (bool, error) {
	for _, ns := range nameservers {
		if addPort {
			ns = net.JoinHostPort(ns, c.authoritativeNSPort)
		}

		result, err := c.lookupTXT(ctx, fqdn, []string{ns}, recursive)
		if err != nil {
			return false, err
		}

		if !matcher(result.Records) {
			return false, fmt.Errorf("NS %s did not return a matching TXT record [fqdn: %s]: %s", ns, fqdn, result)
		}
	}

	return true, nil
}

// lookupAuthoritativeNameservers returns the authoritative nameservers for the given fqdn.
/*
 * NOTE(ldez): This function is a duplication of `lookupAuthoritativeNameservers()` from `dns01/client_nameservers.go`.
 * The 2 functions should be kept in sync.
 */
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
