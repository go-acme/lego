package dnspersist01

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/miekg/dns"
)

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

// getNameservers attempts to get systems nameservers before falling back to the defaults.
/*
 * NOTE(ldez): This function is a duplication of `getNameservers()` from `dns01/client_nameservers.go`.
 * The 2 functions should be kept in sync.
 */
func getNameservers(path string, stack challenge.NetworkStack) []string {
	config, err := dns.ClientConfigFromFile(path)
	if err == nil && len(config.Servers) > 0 {
		return config.Servers
	}

	switch stack {
	case challenge.IPv4Only:
		return []string{
			"1.1.1.1:53",
			"1.0.0.1:53",
		}

	case challenge.IPv6Only:
		return []string{
			"[2606:4700:4700::1111]:53",
			"[2606:4700:4700::1001]:53",
		}

	default:
		return []string{
			"1.1.1.1:53",
			"1.0.0.1:53",
			"[2606:4700:4700::1111]:53",
			"[2606:4700:4700::1001]:53",
		}
	}
}

/*
 * NOTE(ldez): This function is a duplication of `parseNameservers()` from `dns01/client_nameservers.go`.
 * The 2 functions should be kept in sync.
 */
func parseNameservers(servers []string) []string {
	var resolvers []string

	for _, resolver := range servers {
		// ensure all servers have a port number
		if _, _, err := net.SplitHostPort(resolver); err != nil {
			resolvers = append(resolvers, net.JoinHostPort(resolver, "53"))
		} else {
			resolvers = append(resolvers, resolver)
		}
	}

	return resolvers
}
