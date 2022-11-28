package dns01

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// ExtractSubDomain extracts the subdomain part from a domain and a zone.
func ExtractSubDomain(domain, zone string) (string, error) {
	canonDomain := dns.Fqdn(domain)
	canonZone := dns.Fqdn(zone)

	if canonDomain == canonZone {
		return "", fmt.Errorf("no subdomain because the domain and the zone are identical: %s", canonDomain)
	}

	if !dns.IsSubDomain(canonZone, canonDomain) {
		return "", fmt.Errorf("%s is not a subdomain of %s", canonDomain, canonZone)
	}

	return strings.TrimSuffix(canonDomain, "."+canonZone), nil
}
