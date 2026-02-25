package dnspersist01

import (
	"iter"

	"github.com/miekg/dns"
)

/*
 * NOTE(ldez): This file is a duplication of `dns01/fqdn.go`.
 * The 2 files should be kept in sync.
 */

// DomainsSeq generates a sequence of domain names derived from a domain (FQDN or not) in descending order.
func DomainsSeq(fqdn string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if fqdn == "" {
			return
		}

		for _, index := range dns.Split(fqdn) {
			if !yield(fqdn[index:]) {
				return
			}
		}
	}
}
