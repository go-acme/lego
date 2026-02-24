package dnspersist01

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const defaultResolvConf = "/etc/resolv.conf"

// Resolver performs DNS lookups using the configured nameservers and timeout.
type Resolver struct {
	Nameservers []string
	Timeout     time.Duration
}

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

// NewResolver creates a resolver with normalized nameservers and default timeout.
// If nameservers is empty, the system resolv.conf is used, falling back to defaults.
func NewResolver(nameservers []string) *Resolver {
	if len(nameservers) == 0 {
		nameservers = DefaultNameservers()
	}

	return &Resolver{
		Nameservers: ParseNameservers(nameservers),
		Timeout:     DefaultDNSTimeout(),
	}
}

// DefaultNameservers returns resolvers from resolv.conf, falling back to defaults.
func DefaultNameservers() []string {
	config, err := dns.ClientConfigFromFile(defaultResolvConf)
	if err != nil || len(config.Servers) == 0 {
		return defaultFallbackNameservers()
	}

	return ParseNameservers(config.Servers)
}

func defaultFallbackNameservers() []string {
	return []string{
		"google-public-dns-a.google.com:53",
		"google-public-dns-b.google.com:53",
	}
}

// ParseNameservers ensures all servers have a port number.
func ParseNameservers(servers []string) []string {
	var resolvers []string

	for _, resolver := range servers {
		if _, _, err := net.SplitHostPort(resolver); err != nil {
			resolvers = append(resolvers, net.JoinHostPort(resolver, "53"))
		} else {
			resolvers = append(resolvers, resolver)
		}
	}

	return resolvers
}

// LookupTXT resolves TXT records at fqdn. If CNAMEs are returned, they are
// followed up to 50 times to resolve TXT records.
func (r *Resolver) LookupTXT(fqdn string) (TXTResult, error) {
	return r.lookupTXT(fqdn, r.Nameservers, true)
}

func (r *Resolver) lookupTXT(fqdn string, nameservers []string, recursive bool) (TXTResult, error) {
	var result TXTResult

	if r == nil {
		return result, errors.New("resolver is nil")
	}

	nameservers = ParseNameservers(nameservers)
	if len(nameservers) == 0 {
		return result, errors.New("empty list of nameservers")
	}

	timeout := r.Timeout
	if timeout <= 0 {
		timeout = DefaultDNSTimeout()
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

		msg, err := dnsQueryWithTimeout(name, dns.TypeTXT, nameservers, recursive, timeout)
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

func extractCNAME(msg *dns.Msg, name string) string {
	for _, rr := range msg.Answer {
		cn, ok := rr.(*dns.CNAME)
		if !ok {
			continue
		}

		if strings.EqualFold(cn.Hdr.Name, name) {
			return cn.Target
		}
	}

	return ""
}

func dnsQueryWithTimeout(fqdn string, rtype uint16, nameservers []string, recursive bool, timeout time.Duration) (*dns.Msg, error) {
	m := createDNSMsg(fqdn, rtype, recursive)

	if len(nameservers) == 0 {
		return nil, &DNSError{Message: "empty list of nameservers"}
	}

	var (
		msg    *dns.Msg
		err    error
		errAll error
	)

	for _, ns := range nameservers {
		msg, err = sendDNSQuery(m, ns, timeout)
		if err == nil && len(msg.Answer) > 0 {
			break
		}

		errAll = errors.Join(errAll, err)
	}

	if err != nil {
		return msg, errAll
	}

	return msg, nil
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

func sendDNSQuery(m *dns.Msg, ns string, timeout time.Duration) (*dns.Msg, error) {
	if ok, _ := strconv.ParseBool(os.Getenv("LEGO_EXPERIMENTAL_DNS_TCP_ONLY")); ok {
		tcp := &dns.Client{Net: "tcp", Timeout: timeout}

		msg, _, err := tcp.Exchange(m, ns)
		if err != nil {
			return msg, &DNSError{Message: "DNS call error", MsgIn: m, NS: ns, Err: err}
		}

		return msg, nil
	}

	udp := &dns.Client{Net: "udp", Timeout: timeout}
	msg, _, err := udp.Exchange(m, ns)

	if msg != nil && msg.Truncated {
		tcp := &dns.Client{Net: "tcp", Timeout: timeout}
		msg, _, err = tcp.Exchange(m, ns)
	}

	if err != nil {
		return msg, &DNSError{Message: "DNS call error", MsgIn: m, NS: ns, Err: err}
	}

	return msg, nil
}

// DNSError is an error related to DNS calls.
type DNSError struct {
	Message string
	NS      string
	MsgIn   *dns.Msg
	MsgOut  *dns.Msg
	Err     error
}

func (d *DNSError) Error() string {
	var details []string
	if d.NS != "" {
		details = append(details, "ns="+d.NS)
	}

	formatQuestions := func(questions []dns.Question) string {
		var parts []string
		for _, question := range questions {
			parts = append(parts, strings.ReplaceAll(strings.TrimPrefix(question.String(), ";"), "\t", " "))
		}

		return strings.Join(parts, ";")
	}

	if d.MsgIn != nil && len(d.MsgIn.Question) > 0 {
		details = append(details, fmt.Sprintf("question='%s'", formatQuestions(d.MsgIn.Question)))
	}

	if d.MsgOut != nil {
		if d.MsgIn == nil || len(d.MsgIn.Question) == 0 {
			details = append(details, fmt.Sprintf("question='%s'", formatQuestions(d.MsgOut.Question)))
		}

		details = append(details, "code="+dns.RcodeToString[d.MsgOut.Rcode])
	}

	msg := "DNS error"
	if d.Message != "" {
		msg = d.Message
	}

	if d.Err != nil {
		msg += ": " + d.Err.Error()
	}

	if len(details) > 0 {
		msg += " [" + strings.Join(details, ", ") + "]"
	}

	return msg
}

func (d *DNSError) Unwrap() error {
	return d.Err
}
