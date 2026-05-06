package internal

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
)

// FindZoneByFqdn determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func (c *Client) FindZoneByFqdn(ctx context.Context, fqdn string) (string, error) {
	return c.FindZoneByFqdnCustom(ctx, fqdn, c.recursiveNameservers)
}

// FindZoneByFqdnCustom determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func (c *Client) FindZoneByFqdnCustom(ctx context.Context, fqdn string, nameservers []string) (string, error) {
	soa, err := c.lookupSoaByFqdn(ctx, fqdn, nameservers)
	if err != nil {
		return "", fmt.Errorf("[fqdn=%s] %w", fqdn, err)
	}

	return soa.zone, nil
}

func (c *Client) lookupSoaByFqdn(ctx context.Context, fqdn string, nameservers []string) (*soaCacheEntry, error) {
	c.muFqdnSoaCache.Lock()
	defer c.muFqdnSoaCache.Unlock()

	// Do we have it cached and is it still fresh?
	if ent := c.fqdnSoaCache[fqdn]; ent != nil && !ent.isExpired() {
		return ent, nil
	}

	ent, err := c.fetchSoaByFqdn(ctx, fqdn, nameservers)
	if err != nil {
		return nil, err
	}

	c.fqdnSoaCache[fqdn] = ent

	return ent, nil
}

func (c *Client) fetchSoaByFqdn(ctx context.Context, fqdn string, nameservers []string) (*soaCacheEntry, error) {
	var (
		err error
		r   *dns.Msg
	)

	for domain := range DomainsSeq(fqdn) {
		r, err = c.SendQueryCustom(ctx, domain, dns.TypeSOA, nameservers, true)
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
			if msgContainsCNAME(r) {
				continue
			}

			for _, ans := range r.Answer {
				if soa, ok := ans.(*dns.SOA); ok {
					return newSoaCacheEntry(soa), nil
				}
			}
		case dns.RcodeNameError:
			// NXDOMAIN
		default:
			// Any response code other than NOERROR and NXDOMAIN is treated as error
			return nil, &DNSError{Message: fmt.Sprintf("unexpected response for '%s'", domain), MsgOut: r}
		}
	}

	return nil, &DNSError{Message: fmt.Sprintf("could not find the start of authority for '%s'", fqdn), MsgOut: r, Err: err}
}
