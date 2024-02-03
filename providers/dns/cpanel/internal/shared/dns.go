package shared

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

type DNSClient struct {
	timeout time.Duration
}

func NewDNSClient(timeout time.Duration) *DNSClient {
	return &DNSClient{timeout: timeout}
}

func (d DNSClient) SOACall(fqdn, nameserver string) (*dns.SOA, error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeSOA)
	m.SetEdns0(4096, false)

	in, err := d.sendDNSQuery(m, nameserver)
	if err != nil {
		return nil, err
	}

	if len(in.Answer) == 0 {
		if len(in.Ns) > 0 {
			if soa, ok := in.Ns[0].(*dns.SOA); ok && fqdn != soa.Hdr.Name {
				return d.SOACall(soa.Hdr.Name, nameserver)
			}
		}

		return nil, fmt.Errorf("empty answer for %s in %s", fqdn, nameserver)
	}

	for _, rr := range in.Answer {
		if soa, ok := rr.(*dns.SOA); ok {
			return soa, nil
		}
	}

	return nil, fmt.Errorf("SOA not found for %s in %s", fqdn, nameserver)
}

func (d DNSClient) sendDNSQuery(m *dns.Msg, ns string) (*dns.Msg, error) {
	if ok, _ := strconv.ParseBool(os.Getenv("LEGO_EXPERIMENTAL_DNS_TCP_ONLY")); ok {
		tcp := &dns.Client{Net: "tcp", Timeout: d.timeout}
		in, _, err := tcp.Exchange(m, ns)

		return in, err
	}

	udp := &dns.Client{Net: "udp", Timeout: d.timeout}
	in, _, err := udp.Exchange(m, ns)

	if in != nil && in.Truncated {
		tcp := &dns.Client{Net: "tcp", Timeout: d.timeout}
		// If the TCP request succeeds, the err will reset to nil
		in, _, err = tcp.Exchange(m, ns)
	}

	return in, err
}
