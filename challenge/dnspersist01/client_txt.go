package dnspersist01

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// TXTRecord captures a DNS TXT record value and its TTL.
type TXTRecord struct {
	Value string
	TTL   uint32
}

// TXTResult contains TXT records and any CNAMEs followed during lookup.
type TXTResult struct {
	Records    []TXTRecord
	CNAMEChain []string
}

func (r TXTResult) String() string {
	values := make([]string, 0, len(r.Records))

	for _, record := range r.Records {
		values = append(values, record.Value)
	}

	return strings.Join(values, ",")
}

// LookupTXT resolves TXT records at fqdn.
// If CNAMEs are returned, they are followed up to 50 times to resolve TXT records.
func (c *Client) LookupTXT(ctx context.Context, fqdn string) (TXTResult, error) {
	return c.lookupTXT(ctx, fqdn, c.recursiveNameservers, true)
}

func (c *Client) lookupTXT(ctx context.Context, fqdn string, nameservers []string, recursive bool) (TXTResult, error) {
	var result TXTResult

	if c == nil {
		return result, errors.New("resolver is nil")
	}

	nameservers = parseNameservers(nameservers)
	if len(nameservers) == 0 {
		return result, errors.New("empty list of nameservers")
	}

	const maxCNAMEFollows = 50

	name := dns.Fqdn(fqdn)
	seen := map[string]struct{}{}
	followed := 0

	for {
		if _, ok := seen[name]; ok {
			return result, fmt.Errorf("CNAME loop detected for %s", name)
		}

		seen[name] = struct{}{}

		msg, err := c.sendQueryCustom(ctx, name, dns.TypeTXT, nameservers, recursive)
		if err != nil {
			return result, err
		}

		switch msg.Rcode {
		case dns.RcodeSuccess:
			records := extractTXT(msg, name)
			if len(records) > 0 {
				result.Records = records
				return result, nil
			}

			cname := extractCNAME(msg, name)
			if cname == "" {
				return result, nil
			}

			if followed >= maxCNAMEFollows {
				return result, nil
			}

			result.CNAMEChain = append(result.CNAMEChain, cname)
			name = cname
			followed++
		case dns.RcodeNameError:
			return result, nil
		default:
			return result, &DNSError{Message: fmt.Sprintf("unexpected response for '%s'", name), MsgOut: msg}
		}
	}
}

func extractTXT(msg *dns.Msg, name string) []TXTRecord {
	var records []TXTRecord

	for _, rr := range msg.Answer {
		txt, ok := rr.(*dns.TXT)
		if !ok {
			continue
		}

		if !strings.EqualFold(txt.Hdr.Name, name) {
			continue
		}

		records = append(records, TXTRecord{
			Value: strings.Join(txt.Txt, ""),
			TTL:   txt.Hdr.Ttl,
		})
	}

	return records
}
