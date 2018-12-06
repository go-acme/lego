package dns01

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const defaultResolvConf = "/etc/resolv.conf"

// DNSTimeout is used to override the default DNS timeout of 10 seconds.
// FIXME how to define this in Config?
var DNSTimeout = 10 * time.Second

var (
	fqdnToZone   = map[string]string{}
	muFqdnToZone sync.Mutex
)

var defaultNameservers = []string{
	"google-public-dns-a.google.com:53",
	"google-public-dns-b.google.com:53",
}

// RecursiveNameservers are used to pre-check DNS propagation
var RecursiveNameservers = getNameservers(defaultResolvConf, defaultNameservers)

// getNameservers attempts to get systems nameservers before falling back to the defaults
func getNameservers(path string, defaults []string) []string {
	config, err := dns.ClientConfigFromFile(path)
	if err != nil || len(config.Servers) == 0 {
		return defaults
	}

	var systemNameservers []string
	for _, server := range config.Servers {
		// ensure all servers have a port number
		if _, _, err := net.SplitHostPort(server); err != nil {
			systemNameservers = append(systemNameservers, net.JoinHostPort(server, "53"))
		} else {
			systemNameservers = append(systemNameservers, server)
		}
	}
	return systemNameservers
}

// lookupNameservers returns the authoritative nameservers for the given fqdn.
func lookupNameservers(fqdn string) ([]string, error) {
	var authoritativeNss []string

	zone, err := FindZoneByFqdn(fqdn, RecursiveNameservers)
	if err != nil {
		return nil, fmt.Errorf("could not determine the zone: %v", err)
	}

	r, err := dnsQuery(zone, dns.TypeNS, RecursiveNameservers, true)
	if err != nil {
		return nil, err
	}

	for _, rr := range r.Answer {
		if ns, ok := rr.(*dns.NS); ok {
			authoritativeNss = append(authoritativeNss, strings.ToLower(ns.Ns))
		}
	}

	if len(authoritativeNss) > 0 {
		return authoritativeNss, nil
	}
	return nil, fmt.Errorf("could not determine authoritative nameservers")
}

// FindZoneByFqdn determines the zone apex for the given fqdn by recursing up the
// domain labels until the nameserver returns a SOA record in the answer section.
func FindZoneByFqdn(fqdn string, nameservers []string) (string, error) {
	muFqdnToZone.Lock()
	defer muFqdnToZone.Unlock()

	// Do we have it cached?
	if zone, ok := fqdnToZone[fqdn]; ok {
		return zone, nil
	}

	labelIndexes := dns.Split(fqdn)
	for _, index := range labelIndexes {
		domain := fqdn[index:]

		in, err := dnsQuery(domain, dns.TypeSOA, nameservers, true)
		if err != nil {
			return "", err
		}

		// Any response code other than NOERROR and NXDOMAIN is treated as error
		if in.Rcode != dns.RcodeNameError && in.Rcode != dns.RcodeSuccess {
			return "", fmt.Errorf("unexpected response code '%s' for %s",
				dns.RcodeToString[in.Rcode], domain)
		}

		// Check if we got a SOA RR in the answer section
		if in.Rcode == dns.RcodeSuccess {

			// CNAME records cannot/should not exist at the root of a zone.
			// So we skip a domain when a CNAME is found.
			if dnsMsgContainsCNAME(in) {
				continue
			}

			for _, ans := range in.Answer {
				if soa, ok := ans.(*dns.SOA); ok {
					zone := soa.Hdr.Name
					fqdnToZone[fqdn] = zone
					return zone, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find the start of authority")
}

// dnsMsgContainsCNAME checks for a CNAME answer in msg
func dnsMsgContainsCNAME(msg *dns.Msg) bool {
	for _, ans := range msg.Answer {
		if _, ok := ans.(*dns.CNAME); ok {
			return true
		}
	}
	return false
}

// ClearFqdnCache clears the cache of fqdn to zone mappings. Primarily used in testing.
func ClearFqdnCache() {
	fqdnToZone = map[string]string{}
}

// dnsQuery will query a nameserver, iterating through the supplied servers as it retries
// The nameserver should include a port, to facilitate testing where we talk to a mock dns server.
func dnsQuery(fqdn string, rtype uint16, nameservers []string, recursive bool) (in *dns.Msg, err error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, rtype)
	m.SetEdns0(4096, false)

	if !recursive {
		m.RecursionDesired = false
	}

	// Will retry the request based on the number of servers (n+1)
	for i := 1; i <= len(nameservers)+1; i++ {
		ns := nameservers[i%len(nameservers)]
		udp := &dns.Client{Net: "udp", Timeout: DNSTimeout}
		in, _, err = udp.Exchange(m, ns)

		if err == dns.ErrTruncated {
			tcp := &dns.Client{Net: "tcp", Timeout: DNSTimeout}
			// If the TCP request succeeds, the err will reset to nil
			in, _, err = tcp.Exchange(m, ns)
		}

		if err == nil {
			break
		}
	}
	return
}
