package dns01

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v5/challenge/internal"
	"github.com/go-acme/lego/v5/log"
	"github.com/miekg/dns"
)

func (c *Client) resolveCNAME(ctx context.Context, fqdn string) (string, error) {
	r, err := c.core.SendQuery(ctx, fqdn, dns.TypeTXT, true)
	if err != nil {
		return "", fmt.Errorf("initial recursive nameserver: %w", err)
	}

	if r.Rcode == dns.RcodeSuccess {
		fqdn = updateDomainWithCName(r, fqdn)
	}

	return fqdn, nil
}

func (c *Client) lookupCNAME(ctx context.Context, fqdn string) string {
	// recursion counter so it doesn't spin out of control
	for range 50 {
		// Keep following CNAMEs
		r, err := c.core.SendQuery(ctx, fqdn, dns.TypeCNAME, true)
		if err != nil {
			log.Debug("Lookup CNAME.",
				slog.String("fqdn", fqdn),
				log.ErrorAttr(err),
			)

			// No more CNAME records to follow, exit
			break
		}

		if r.Rcode != dns.RcodeSuccess {
			// No more CNAME records to follow, exit
			break
		}

		// Check if the domain has CNAME then use that
		cname := updateDomainWithCName(r, fqdn)
		if cname == fqdn {
			break
		}

		log.Info("Found CNAME entry.",
			slog.String("fqdn", fqdn),
			slog.String("cname", cname),
		)

		fqdn = cname
	}

	return fqdn
}

// Update FQDN with CNAME if any.
func updateDomainWithCName(r *dns.Msg, fqdn string) string {
	cname := internal.ExtractCNAME(r, fqdn)
	if cname != "" {
		return cname
	}

	return fqdn
}
