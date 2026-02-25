package dnspersist01

import (
	"slices"
	"strings"

	"github.com/miekg/dns"
)

/*
 * NOTE(ldez): This function is a partial duplication of `updateDomainWithCName()` from `dns01/client_cname.go`.
 * The 2 functions should be kept in sync.
 */
func extractCNAME(msg *dns.Msg, name string) string {
	for _, rr := range msg.Answer {
		cn, ok := rr.(*dns.CNAME)
		if !ok {
			continue
		}

		if strings.EqualFold(cn.Hdr.Name, name) {
			return cn.Target
		}
	}

	return ""
}

// dnsMsgContainsCNAME checks for a CNAME answer in msg.
/*
 * NOTE(ldez): This function is a duplication of `Client.sendQuery()` from `dns01/client_cname.go`.
 * The 2 functions should be kept in sync.
 */
func dnsMsgContainsCNAME(msg *dns.Msg) bool {
	return slices.ContainsFunc(msg.Answer, func(rr dns.RR) bool {
		_, ok := rr.(*dns.CNAME)
		return ok
	})
}
