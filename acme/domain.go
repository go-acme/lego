package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/publicsuffix"
)

var recursiveNameserver = "google-public-dns-a.google.com:53"

// dnsQuery sends a DNS query to the given nameserver.
// The nameserver should include a port, to facilitate testing where we talk to a mock dns server.
func DnsQuery(fqdn string, rtype uint16, nameserver string, recursive bool) (in *dns.Msg, err error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, rtype)
	m.SetEdns0(4096, false)

	if !recursive {
		m.RecursionDesired = false
	}

	in, err = dns.Exchange(m, nameserver)
	if err == dns.ErrTruncated {
		tcp := &dns.Client{Net: "tcp"}
		in, _, err = tcp.Exchange(m, nameserver)
	}

	return
}

// toFqdn converts the name into a fqdn appending a trailing dot.
func ToFqdn(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	}
	return name + "."
}

// unFqdn converts the fqdn into a name removing the trailing dot.
func UnFqdn(name string) string {
	n := len(name)
	if n != 0 && name[n-1] == '.' {
		return name[:n-1]
	}
	return name
}


type Domain struct {
	Domain                string
	authoritativeZone     string
	nameServers       	  []string
}

func NewDomain(domain string) *Domain {
	return &Domain{Domain: ToFqdn(domain)}
}

// DNS01Record returns a DNS record which will fulfill the `dns-01` challenge
func (d *Domain) GetDNS01Record(keyAuth string) (fqdn string, value string, ttl int) {
	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// base64URL encoding without padding
	keyAuthSha := base64.URLEncoding.EncodeToString(keyAuthShaBytes[:sha256.Size])
	value = strings.TrimRight(keyAuthSha, "=")
	ttl = 120
	fqdn = fmt.Sprintf("_acme-challenge.%s", d.Domain)
	return
}

// GetFqdn returns the fqdn form of the record name
func (d *Domain) GetFqdn() string {
	return d.Domain
}

// GetUnFqdn returns the non-fqdn form of the record name
func (d *Domain) GetUnFqdn() string {
	return UnFqdn(d.Domain)
}

// GetAuthoritativeZone determines the authoritative zone of the given fqdn
func (d *Domain) GetAuthoritativeZone() (string, error) {

	// Do we have it cached?
	if len(d.authoritativeZone) > 0 {
		return d.authoritativeZone, nil
	}

	// Query the authorative nameserver for a hopefully non-existing SOA record,
	// in the authority section of the reply it will have the SOA of the
	// containing zone. rfc2308 has this to say on the subject:
	//   Name servers authoritative for a zone MUST include the SOA record of
	//   the zone in the authority section of the response when reporting an
	//   NXDOMAIN or indicating that no data (NODATA) of the requested type exists
	in, err := DnsQuery(d.GetFqdn(), dns.TypeSOA, recursiveNameserver, true)
	if err != nil {
		return "", err
	}
	if in.Rcode != dns.RcodeNameError {
		if in.Rcode != dns.RcodeSuccess {
			return "", fmt.Errorf("NS %s returned %s for %s", recursiveNameserver, dns.RcodeToString[in.Rcode], d.Domain)
		}
		// We have a success, so one of the answers has to be a SOA RR
		for _, ans := range in.Answer {
			if soa, ok := ans.(*dns.SOA); ok {
				zone := soa.Hdr.Name
				// If we ended up on one of the TLDs, it means the domain did not exist.
				publicsuffix, _ := publicsuffix.PublicSuffix(UnFqdn(zone))
				if publicsuffix == UnFqdn(zone) {
					return "", fmt.Errorf("Could not determine zone authoritatively")
				}

				d.authoritativeZone = zone
				return d.authoritativeZone, nil
			}
		}
		// Or it is NODATA, fall through to NXDOMAIN
	}
	// Search the authority section for our precious SOA RR
	for _, ns := range in.Ns {
		if soa, ok := ns.(*dns.SOA); ok {
			zone := soa.Hdr.Name
			// If we ended up on one of the TLDs, it means the domain did not exist.
			publicsuffix, _ := publicsuffix.PublicSuffix(UnFqdn(zone))
			if publicsuffix == UnFqdn(zone) {
				return "", fmt.Errorf("Could not determine zone authoritatively")
			}

			d.authoritativeZone = zone
			return d.authoritativeZone, nil
		}
	}
	return "", fmt.Errorf("NS %s did not return the expected SOA record in the authority section", recursiveNameserver)
}

// CheckDNSPropagation checks if the expected TXT record has been propagated to all authoritative nameservers.
func (d *Domain) CheckDNSPropagation(value string) (bool, error) {

	fqdn := d.GetFqdn()

	// Initial attempt to resolve at the recursive NS
	r, err := DnsQuery(fqdn, dns.TypeTXT, recursiveNameserver, true)
	if err != nil {
		return false, err
	}
	if r.Rcode == dns.RcodeSuccess {
		// If we see a CNAME here then use the alias
		for _, rr := range r.Answer {
			if cn, ok := rr.(*dns.CNAME); ok {
				if cn.Hdr.Name == fqdn {
					fqdn = cn.Target
					break
				}
			}
		}
	}

	authoritativeNss, err := d.LookupNameservers()
	if err != nil {
		return false, err
	}

	return checkAuthoritativeNss(fqdn, value, authoritativeNss, "53")
}

// LookupNameservers returns the authoritative nameservers for the given fqdn.
func (d *Domain) LookupNameservers() ([]string, error) {
	var authoritativeNss []string

	if len(d.nameServers) > 0 {
		return d.nameServers, nil
	}

	zone, err := d.GetAuthoritativeZone()
	if err != nil {
		return nil, err
	}

	r, err := DnsQuery(zone, dns.TypeNS, recursiveNameserver, true)
	if err != nil {
		return nil, err
	}

	for _, rr := range r.Answer {
		if ns, ok := rr.(*dns.NS); ok {
			authoritativeNss = append(authoritativeNss, ns.Ns)
		}
	}

	if len(authoritativeNss) > 0 {
		d.nameServers = authoritativeNss
		return authoritativeNss, nil
	}
	return nil, fmt.Errorf("Could not determine authoritative nameservers")
}

// checkAuthoritativeNss queries each of the given nameservers for the expected TXT record.
func checkAuthoritativeNss(fqdn, value string, nameservers []string, port string) (bool, error) {
	for _, ns := range nameservers {
		r, err := DnsQuery(fqdn, dns.TypeTXT, net.JoinHostPort(ns, port), false)
		if err != nil {
			return false, err
		}

		if r.Rcode != dns.RcodeSuccess {
			return false, fmt.Errorf("NS %s returned %s for %s", ns, dns.RcodeToString[r.Rcode], fqdn)
		}

		var found bool
		for _, rr := range r.Answer {
			if txt, ok := rr.(*dns.TXT); ok {
				if strings.Join(txt.Txt, "") == value {
					found = true
					break
				}
			}
		}

		if !found {
			return false, fmt.Errorf("NS %s did not return the expected TXT record", ns)
		}
	}

	return true, nil
}
