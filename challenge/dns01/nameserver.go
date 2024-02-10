package dns01

import (
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const defaultResolvConf = "/etc/resolv.conf"

var (
	fqdnSoaCache   = map[string]*soaCacheEntry{}
	muFqdnSoaCache sync.Mutex
)

var defaultNameservers = []string{
	"google-public-dns-a.google.com:53",
	"google-public-dns-b.google.com:53",
}

// recursiveNameservers are used to pre-check DNS propagation.
var recursiveNameservers = getNameservers(defaultResolvConf, defaultNameservers)

// soaCacheEntry holds a cached SOA record (only selected fields).
type soaCacheEntry struct {
	zone      string    // zone apex (a domain name)
	primaryNs string    // primary nameserver for the zone apex
	expires   time.Time // time when this cache entry should be evicted
}

func newSoaCacheEntry(soa *dns.SOA) *soaCacheEntry {
	return &soaCacheEntry{
		zone:      soa.Hdr.Name,
		primaryNs: soa.Ns,
		expires:   time.Now().Add(time.Duration(soa.Refresh) * time.Second),
	}
}

// isExpired checks whether a cache entry should be considered expired.
func (cache *soaCacheEntry) isExpired() bool {
	return time.Now().After(cache.expires)
}

// ClearFqdnCache clears the cache of fqdn to zone mappings. Primarily used in testing.
func ClearFqdnCache() {
	muFqdnSoaCache.Lock()
	fqdnSoaCache = map[string]*soaCacheEntry{}
	muFqdnSoaCache.Unlock()
}

func AddDNSTimeout(timeout time.Duration) ChallengeOption {
	return func(_ *Challenge) error {
		dnsTimeout = timeout
		return nil
	}
}

func AddRecursiveNameservers(nameservers []string) ChallengeOption {
	return func(_ *Challenge) error {
		recursiveNameservers = ParseNameservers(nameservers)
		return nil
	}
}

// getNameservers attempts to get systems nameservers before falling back to the defaults.
func getNameservers(path string, defaults []string) []string {
	config, err := dns.ClientConfigFromFile(path)
	if err != nil || len(config.Servers) == 0 {
		return defaults
	}

	return ParseNameservers(config.Servers)
}

func ParseNameservers(servers []string) []string {
	var resolvers []string
	for _, resolver := range servers {
		// ensure all servers have a port number
		if _, _, err := net.SplitHostPort(resolver); err != nil {
			resolvers = append(resolvers, net.JoinHostPort(resolver, "53"))
		} else {
			resolvers = append(resolvers, resolver)
		}
	}
	return resolvers
}

// lookupNameservers returns the authoritative nameservers for the given fqdn.
func lookupNameservers(fqdn string) ([]string, error) {
	var authoritativeNss []string

	zone, err := FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, fmt.Errorf("[fqdn=%s] could not determine the zone: %w", fqdn, err)
	}

	r, err := dnsQuery(zone, dns.TypeNS, recursiveNameservers, true)
	if err != nil {
		return nil, fmt.Errorf("[zone=%s] NS call failed: %w", zone, err)
	}

	for _, rr := range r.Answer {
		if ns, ok := rr.(*dns.NS); ok {
			authoritativeNss = append(authoritativeNss, strings.ToLower(ns.Ns))
		}
	}

	if len(authoritativeNss) > 0 {
		return authoritativeNss, nil
	}

	return nil, fmt.Errorf("[zone=%s] could not determine authoritative nameservers", zone)
}

// FindPrimaryNsByFqdn determines the primary nameserver of the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func FindPrimaryNsByFqdn(fqdn string) (string, error) {
	return FindPrimaryNsByFqdnCustom(fqdn, recursiveNameservers)
}

// FindPrimaryNsByFqdnCustom determines the primary nameserver of the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func FindPrimaryNsByFqdnCustom(fqdn string, nameservers []string) (string, error) {
	soa, err := lookupSoaByFqdn(fqdn, nameservers)
	if err != nil {
		return "", err
	}
	return soa.primaryNs, nil
}

// FindZoneByFqdn determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func FindZoneByFqdn(fqdn string) (string, error) {
	return FindZoneByFqdnCustom(fqdn, recursiveNameservers)
}

// FindZoneByFqdnCustom determines the zone apex for the given fqdn
// by recursing up the domain labels until the nameserver returns a SOA record in the answer section.
func FindZoneByFqdnCustom(fqdn string, nameservers []string) (string, error) {
	soa, err := lookupSoaByFqdn(fqdn, nameservers)
	if err != nil {
		return "", err
	}
	return soa.zone, nil
}

func lookupSoaByFqdn(fqdn string, nameservers []string) (*soaCacheEntry, error) {
	muFqdnSoaCache.Lock()
	defer muFqdnSoaCache.Unlock()

	// Do we have it cached and is it still fresh?
	if ent := fqdnSoaCache[fqdn]; ent != nil && !ent.isExpired() {
		return ent, nil
	}

	ent, err := fetchSoaByFqdn(fqdn, nameservers)
	if err != nil {
		return nil, err
	}

	fqdnSoaCache[fqdn] = ent
	return ent, nil
}

func fetchSoaByFqdn(fqdn string, nameservers []string) (*soaCacheEntry, error) {
	var err error
	var in *dns.Msg

	labelIndexes := dns.Split(fqdn)
	for _, index := range labelIndexes {
		domain := fqdn[index:]

		in, err = dnsQuery(domain, dns.TypeSOA, nameservers, true)
		if err != nil {
			continue
		}

		if in == nil {
			continue
		}

		switch in.Rcode {
		case dns.RcodeSuccess:
			// Check if we got a SOA RR in the answer section
			if len(in.Answer) == 0 {
				continue
			}

			// CNAME records cannot/should not exist at the root of a zone.
			// So we skip a domain when a CNAME is found.
			if dnsMsgContainsCNAME(in) {
				continue
			}

			for _, ans := range in.Answer {
				if soa, ok := ans.(*dns.SOA); ok {
					return newSoaCacheEntry(soa), nil
				}
			}
		case dns.RcodeNameError:
			// NXDOMAIN
		default:
			// Any response code other than NOERROR and NXDOMAIN is treated as error
			return nil, fmt.Errorf("unexpected response code '%s' for %s", dns.RcodeToString[in.Rcode], domain)
		}
	}

	errMsg := "could not find the start of authority for %s"
	if errW := wrapDNSError(in, err); errW != nil {
		return nil, fmt.Errorf(errMsg+": %w", fqdn, errW)
	}

	return nil, fmt.Errorf(errMsg, fqdn)
}

// dnsMsgContainsCNAME checks for a CNAME answer in msg.
func dnsMsgContainsCNAME(msg *dns.Msg) bool {
	return slices.ContainsFunc(msg.Answer, func(rr dns.RR) bool {
		_, ok := rr.(*dns.CNAME)
		return ok
	})
}

func dnsQuery(fqdn string, rtype uint16, nameservers []string, recursive bool) (*dns.Msg, error) {
	m := createDNSMsg(fqdn, rtype, recursive)

	var in *dns.Msg
	var err error
	var errAll error

	for _, ns := range nameservers {
		in, err = sendDNSQuery(m, ns)
		if err == nil && len(in.Answer) > 0 {
			break
		}

		errAll = errors.Join(errAll, err)
	}

	if err != nil {
		return in, errAll
	}

	return in, nil
}

func createDNSMsg(fqdn string, rtype uint16, recursive bool) *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, rtype)
	m.SetEdns0(4096, false)

	if !recursive {
		m.RecursionDesired = false
	}

	return m
}

func sendDNSQuery(m *dns.Msg, ns string) (*dns.Msg, error) {
	if ok, _ := strconv.ParseBool(os.Getenv("LEGO_EXPERIMENTAL_DNS_TCP_ONLY")); ok {
		tcp := &dns.Client{Net: "tcp", Timeout: dnsTimeout}
		in, _, err := tcp.Exchange(m, ns)
		if err != nil {
			return in, fmt.Errorf("DNS call to %s: %w", ns, err)
		}

		return in, nil
	}

	udp := &dns.Client{Net: "udp", Timeout: dnsTimeout}
	in, _, err := udp.Exchange(m, ns)

	if in != nil && in.Truncated {
		tcp := &dns.Client{Net: "tcp", Timeout: dnsTimeout}
		// If the TCP request succeeds, the "err" will reset to nil
		in, _, err = tcp.Exchange(m, ns)
	}

	if err != nil {
		return in, fmt.Errorf("DNS call to %s: %w", ns, err)
	}

	return in, nil
}

func wrapDNSError(msg *dns.Msg, err error) error {
	if msg != nil {
		m := fmt.Sprintf("code=%s, question='%s'", dns.RcodeToString[msg.Rcode], formatQuestions(msg.Question))
		if err != nil {
			return fmt.Errorf("%s: %w", m, err)
		}

		return errors.New(m)
	}

	return err
}

func formatQuestions(questions []dns.Question) string {
	var parts []string
	for _, question := range questions {
		parts = append(parts, strings.ReplaceAll(strings.TrimPrefix(question.String(), ";"), "\t", " "))
	}

	return strings.Join(parts, ";")
}
