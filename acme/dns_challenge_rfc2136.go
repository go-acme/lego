package acme

import (
	"fmt"
	"github.com/miekg/dns"
	"time"
)

// DNSProviderRFC2136 is an implementation of the DNSProvider interface that
// uses dynamic DNS updates (RFC 2136) to create TXT records on a nameserver.
type DNSProviderRFC2136 struct {
	nameserver string
	zone       string
	tsigKey    string
	tsigSecret string
}

// NewDNSProviderRFC2136 returns a new DNSProviderRFC2136 instance.
// To disable TSIG authentication 'tsigKey' and 'tsigSecret' must be set to the empty string.
// 'nameserver' must be a network address in the the form "host:port". 'zone' must be the fully
// qualified name of the zone.
func NewDNSProviderRFC2136(nameserver, zone, tsigKey, tsigSecret string) (*DNSProviderRFC2136, error) {
	d := &DNSProviderRFC2136{
		nameserver: nameserver,
		zone:       zone,
	}
	if len(tsigKey) > 0 && len(tsigSecret) > 0 {
		d.tsigKey = tsigKey
		d.tsigSecret = tsigSecret
	}

	return d, nil
}

// CreateTXTRecord creates a TXT record using the specified parameters
func (r *DNSProviderRFC2136) CreateTXTRecord(fqdn, value string, ttl int) error {
	return r.changeRecord("INSERT", fqdn, value, ttl)
}

// RemoveTXTRecord removes the TXT record matching the specified parameters
func (r *DNSProviderRFC2136) RemoveTXTRecord(fqdn, value string, ttl int) error {
	return r.changeRecord("REMOVE", fqdn, value, ttl)
}

func (r *DNSProviderRFC2136) changeRecord(action, fqdn, value string, ttl int) error {
	// Create RR
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: fqdn, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: uint32(ttl)}
	rr.Txt = []string{value}
	rrs := make([]dns.RR, 1)
	rrs[0] = rr

	// Create dynamic update packet
	m := new(dns.Msg)
	m.SetUpdate(dns.Fqdn(r.zone))
	switch action {
	case "INSERT":
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
		m.SetTsig(dns.Fqdn(r.tsigKey), dns.HmacMD5, 300, time.Now().Unix())
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
