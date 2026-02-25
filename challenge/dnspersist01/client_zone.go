package dnspersist01

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/challenge/internal"
	"github.com/miekg/dns"
)

// FindZoneByFqdn determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
/*
 * NOTE(ldez): This function is a partial duplication of `Client.FindZoneByFqdn()` from `dns01/client_zone.go`.
 * The 2 functions should be kept in sync.
 */
func (c *Client) FindZoneByFqdn(ctx context.Context, fqdn string) (string, error) {
	return c.FindZoneByFqdnCustom(ctx, fqdn, c.recursiveNameservers)
}

// FindZoneByFqdnCustom determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
/*
 * NOTE(ldez): This function is a partial duplication of `Client.FindZoneByFqdnCustom()` from `dns01/client_zone.go`.
 * The 2 functions should be kept in sync.
 */
func (c *Client) FindZoneByFqdnCustom(ctx context.Context, fqdn string, nameservers []string) (string, error) {
	soa, err := c.fetchSoaByFqdn(ctx, fqdn, nameservers)
	if err != nil {
		return "", fmt.Errorf("[fqdn=%s] %w", fqdn, err)
	}

	return soa, nil
}

/*
 * NOTE(ldez): This function is a partial duplication of `Client.fetchSoaByFqdn()` from `dns01/client_zone.go`.
 * The 2 functions should be kept in sync.
 */
func (c *Client) fetchSoaByFqdn(ctx context.Context, fqdn string, nameservers []string) (string, error) {
	var (
		err error
		r   *dns.Msg
	)

	for domain := range domainsSeq(fqdn) {
		r, err = c.sendQueryCustom(ctx, domain, dns.TypeSOA, nameservers, true)
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
			if internal.MsgContainsCNAME(r) {
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

	return "", &DNSError{Message: fmt.Sprintf("could not find the start of authority for '%s'", fqdn), MsgOut: r, Err: err}
}
