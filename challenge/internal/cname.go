package internal

import (
	"slices"
	"strings"

	"github.com/miekg/dns"
)

// ExtractCNAME extracts the CNAME target for a given name.
func ExtractCNAME(msg *dns.Msg, name string) string {
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

// msgContainsCNAME checks for a CNAME answer in msg.
func msgContainsCNAME(msg *dns.Msg) bool {
	return slices.ContainsFunc(msg.Answer, func(rr dns.RR) bool {
		_, ok := rr.(*dns.CNAME)
		return ok
	})
}
