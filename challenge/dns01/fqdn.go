package dns01

import (
	"iter"
	"strings"

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

// isKnownNewerGTLD checks if a domain uses a newer gTLD that should be treated
// as a complete zone rather than being split further.
func isKnownNewerGTLD(domain string) bool {
	// Known newer gTLDs that are commonly used as zones
	// This list can be expanded as needed
	knownNewerGTLDs := []string{
		".dog", ".cat", ".horse", ".wiki", ".blog", ".app", ".dev",
		".page", ".site", ".tech", ".online", ".store", ".cloud",
		".rocks", ".space", ".world", ".fun", ".run", ".live",
	}
	
	domainLower := strings.ToLower(domain)
	for _, tld := range knownNewerGTLDs {
		if strings.HasSuffix(domainLower, tld) || strings.HasSuffix(domainLower, tld+".") {
			return true
		}
	}
	
	return false
}

// DomainsSeq generates a sequence of domain names derived from a domain (FQDN or not) in descending order.
// It includes special handling for newer gTLDs to prevent oversplitting during zone detection.
// 
// For example, "play.app4.dog" will generate ["play.app4.dog", "app4.dog"] instead of 
// ["play.app4.dog", "app4.dog", "dog"], since "dog" is not a valid zone for "app4.dog" domains.
func DomainsSeq(fqdn string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if fqdn == "" {
			return
		}

		for _, index := range dns.Split(fqdn) {
			candidate := fqdn[index:]
			if !yield(candidate) {
				return
			}
			
			// For newer gTLDs, stop splitting when we reach the likely zone
			if isKnownNewerGTLD(candidate) {
				cleanDomain := strings.TrimSuffix(candidate, ".")
				if strings.Count(cleanDomain, ".") == 1 {
					return
				}
			}
		}
	}
}
