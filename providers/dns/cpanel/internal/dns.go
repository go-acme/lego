package internal

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/log"
	"github.com/miekg/dns"
)

type DNSClient struct {
	timeout time.Duration
}

func NewDNSClient(timeout time.Duration) *DNSClient {
	return &DNSClient{timeout: timeout}
}

func (d DNSClient) SOACall(fqdn, ns string) (*dns.SOA, error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeSOA)
	m.SetEdns0(4096, false)

	in, err := d.sendDNSQuery(m, ns)
	if err != nil {
		return nil, err
	}

	log.Println(in)

	if len(in.Answer) <= 0 {
		if len(in.Ns) > 0 {
			name := in.Ns[0].(*dns.SOA).Hdr.Name
			fmt.Println(fqdn != name)
			if fqdn != name {
				return d.SOACall(name, ns)
			}
		}

		return nil, fmt.Errorf("empty answer for %s in %s", fqdn, ns)
	}

	for _, rr := range in.Answer {
		if soa, ok := rr.(*dns.SOA); ok {
			return soa, nil
		}
	}

	return nil, fmt.Errorf("SOA not found for %s in %s", fqdn, ns)
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
