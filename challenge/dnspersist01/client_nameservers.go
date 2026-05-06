package dnspersist01

import (
	"context"
	"fmt"
	"net"
)

// checkRecursiveNameserversPropagation queries each of the recursive nameservers for the expected TXT record.
func (c *Client) checkRecursiveNameserversPropagation(ctx context.Context, fqdn string, matcher RecordMatcher) (bool, error) {
	return c.checkNameserversPropagationCustom(ctx, fqdn, c.core.GetRecursiveNameservers(), matcher, false, true)
}

// checkRecursiveNameserversPropagation queries each of the authoritative nameservers for the expected TXT record.
func (c *Client) checkAuthoritativeNameserversPropagation(ctx context.Context, fqdn string, matcher RecordMatcher) (bool, error) {
	authoritativeNss, err := c.core.LookupAuthoritativeNameservers(ctx, fqdn)
	if err != nil {
		return false, err
	}

	return c.checkNameserversPropagationCustom(ctx, fqdn, authoritativeNss, matcher, true, false)
}

func (c *Client) checkNameserversPropagationCustom(ctx context.Context, fqdn string, nameservers []string, matcher RecordMatcher, addPort, recursive bool) (bool, error) {
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
