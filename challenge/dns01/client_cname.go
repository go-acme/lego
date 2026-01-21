package dns01

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/go-acme/lego/v5/log"
	"github.com/miekg/dns"
)

func (c *Client) lookupCNAME(ctx context.Context, fqdn string) string {
	// recursion counter so it doesn't spin out of control
	for range 50 {
		// Keep following CNAMEs
		r, err := c.sendQuery(ctx, fqdn, dns.TypeCNAME, true)
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
	for _, rr := range r.Answer {
		if cn, ok := rr.(*dns.CNAME); ok {
			if strings.EqualFold(cn.Hdr.Name, fqdn) {
				return cn.Target
			}
		}
	}

	return fqdn
}

// dnsMsgContainsCNAME checks for a CNAME answer in msg.
func dnsMsgContainsCNAME(msg *dns.Msg) bool {
	return slices.ContainsFunc(msg.Answer, func(rr dns.RR) bool {
		_, ok := rr.(*dns.CNAME)
		return ok
	})
}
