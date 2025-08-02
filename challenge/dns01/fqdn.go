package dns01

import (
	"iter"

	"github.com/miekg/dns"
)

// ToFqdn converts the name into a fqdn appending a trailing dot.
//
// Deprecated: Use [github.com/miekg/dns.Fqdn] directly.
func ToFqdn(name string) string {
	return dns.Fqdn(name)
}

// UnFqdn converts the fqdn into a name removing the trailing dot.
func UnFqdn(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}

// UnFqdnDomainsSeq generates a sequence of "unFQDNed" domain names derived from a domain (FQDN or not) in descending order.
func UnFqdnDomainsSeq(fqdn string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if fqdn == "" {
			return
		}

		for _, index := range dns.Split(fqdn) {
			if !yield(UnFqdn(fqdn[index:])) {
				return
			}
		}
	}
}

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
