package acme

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSProviderRFC2136 is an implementation of the ChallengeProvider interface that
// uses dynamic DNS updates (RFC 2136) to create TXT records on a nameserver.
type DNSProviderRFC2136 struct {
	nameserver    string
	tsigAlgorithm string
	tsigKey       string
	tsigSecret    string
	domain2zone   map[string]string
}

// NewDNSProviderRFC2136 returns a new DNSProviderRFC2136 instance.
// To disable TSIG authentication 'tsigAlgorithm, 'tsigKey' and 'tsigSecret' must be set to the empty string.
// 'nameserver' must be a network address in the the form "host" or "host:port".
func NewDNSProviderRFC2136(nameserver, tsigAlgorithm, tsigKey, tsigSecret string) (*DNSProviderRFC2136, error) {
	// Append the default DNS port if none is specified.
	if !strings.Contains(nameserver, ":") {
		nameserver += ":53"
	}
	d := &DNSProviderRFC2136{
		nameserver:  nameserver,
		domain2zone: make(map[string]string),
	}
	if tsigAlgorithm == "" {
		tsigAlgorithm = dns.HmacMD5
	}
	d.tsigAlgorithm = tsigAlgorithm
	if len(tsigKey) > 0 && len(tsigSecret) > 0 {
		d.tsigKey = tsigKey
		d.tsigSecret = tsigSecret
	}

	return d, nil
}

// Present creates a TXT record using the specified parameters
func (r *DNSProviderRFC2136) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := DNS01Record(domain, keyAuth)
	return r.changeRecord("INSERT", fqdn, value, ttl)
}

// CleanUp removes the TXT record matching the specified parameters
func (r *DNSProviderRFC2136) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, ttl := DNS01Record(domain, keyAuth)
	return r.changeRecord("REMOVE", fqdn, value, ttl)
}

func (r *DNSProviderRFC2136) changeRecord(action, fqdn, value string, ttl int) error {
	zone, err := r.findZone(fqdn)
	if err != nil {
		return err
	}

	// Create RR
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: fqdn, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: uint32(ttl)}
	rr.Txt = []string{value}
	rrs := make([]dns.RR, 1)
	rrs[0] = rr

	// Create dynamic update packet
	m := new(dns.Msg)
	m.SetUpdate(zone)
	switch action {
	case "INSERT":
		// Always remove old challenge left over from who knows what.
		m.RemoveRRset(rrs)
		m.Insert(rrs)
	case "REMOVE":
		m.Remove(rrs)
	default:
		return fmt.Errorf("Unexpected action: %s", action)
	}

	// Setup client
	c := new(dns.Client)
	c.SingleInflight = true
	// TSIG authentication / msg signing
	if len(r.tsigKey) > 0 && len(r.tsigSecret) > 0 {
		m.SetTsig(dns.Fqdn(r.tsigKey), r.tsigAlgorithm, 300, time.Now().Unix())
		c.TsigSecret = map[string]string{dns.Fqdn(r.tsigKey): r.tsigSecret}
	}

	// Send the query
	reply, _, err := c.Exchange(m, r.nameserver)
	if err != nil {
		return fmt.Errorf("DNS update failed: %v", err)
	}
	if reply != nil && reply.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update failed. Server replied: %s", dns.RcodeToString[reply.Rcode])
	}

	return nil
}

// findZone determines the zone of a qualifying cert DNSname
func (r *DNSProviderRFC2136) findZone(fqdn string) (string, error) {
	// Do we have it cached?
	if val, ok := r.domain2zone[fqdn]; ok {
		return val, nil
	}

	// Query the authorative nameserver for a hopefully non-existing SOA record,
	// in the authority section of the reply it will have the SOA of the
	// containing zone. rfc2308 has this to say on the subject:
	//   Name servers authoritative for a zone MUST include the SOA record of
	//   the zone in the authority section of the response when reporting an
	//   NXDOMAIN or indicating that no data (NODATA) of the requested type exists
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeSOA)
	m.SetEdns0(4096, false)
	m.RecursionDesired = true
	m.Authoritative = true

	in, err := dns.Exchange(m, r.nameserver)
	if err == dns.ErrTruncated {
		tcp := &dns.Client{Net: "tcp"}
		in, _, err = tcp.Exchange(m, r.nameserver)
	}
	if err != nil {
		return "", err
	}
	if in.Rcode != dns.RcodeNameError {
		if in.Rcode != dns.RcodeSuccess {
			return "", fmt.Errorf("DNS Query for zone %q failed", fqdn)
		}
		// We have a success, so one of the answers has to be a SOA RR
		for _, ans := range in.Answer {
			if ans.Header().Rrtype == dns.TypeSOA {
				zone := ans.Header().Name
				r.domain2zone[fqdn] = zone
				return zone, nil
			}
		}
		// Or it is NODATA, fall through to NXDOMAIN
	}
	// Search the authority section for our precious SOA RR
	for _, ns := range in.Ns {
		if ns.Header().Rrtype == dns.TypeSOA {
			zone := ns.Header().Name
			r.domain2zone[fqdn] = zone
			return zone, nil
		}
	}
	return "", fmt.Errorf("Expected a SOA record in the authority section")
}
