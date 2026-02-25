package dnspersist01

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/miekg/dns"
)

const defaultResolvConf = "/etc/resolv.conf"

var defaultClient atomic.Pointer[Client]

func init() {
	defaultClient.Store(NewClient(nil))
}

func DefaultClient() *Client { return defaultClient.Load() }

func SetDefaultClient(c *Client) {
	defaultClient.Store(c)
}

type Options struct {
	RecursiveNameservers []string
	Timeout              time.Duration
	TCPOnly              bool
	NetworkStack         challenge.NetworkStack
}

type Client struct {
	recursiveNameservers []string

	// authoritativeNSPort used by authoritative NS.
	// For testing purposes only.
	authoritativeNSPort string

	tcpClient *dns.Client
	udpClient *dns.Client
	tcpOnly   bool
}

func NewClient(opts *Options) *Client {
	if opts == nil {
		tcpOnly, _ := strconv.ParseBool(os.Getenv("LEGO_EXPERIMENTAL_DNS_TCP_ONLY"))
		opts = &Options{TCPOnly: tcpOnly}
	}

	if len(opts.RecursiveNameservers) == 0 {
		opts.RecursiveNameservers = getNameservers(defaultResolvConf, opts.NetworkStack)
	}

	if opts.Timeout == 0 {
		opts.Timeout = dnsTimeout
	}

	return &Client{
		recursiveNameservers: parseNameservers(opts.RecursiveNameservers),
		authoritativeNSPort:  "53",
		tcpClient: &dns.Client{
			Net:     opts.NetworkStack.Network("tcp"),
			Timeout: opts.Timeout,
		},
		udpClient: &dns.Client{
			Net:     opts.NetworkStack.Network("udp"),
			Timeout: opts.Timeout,
		},
		tcpOnly: opts.TCPOnly,
	}
}

/*
 * NOTE(ldez): This function is a duplication of `Client.sendQuery()` from `dns01/client.go`.
 * The 2 functions should be kept in sync.
 */
func (c *Client) sendQuery(ctx context.Context, fqdn string, rtype uint16, recursive bool) (*dns.Msg, error) {
	return c.sendQueryCustom(ctx, fqdn, rtype, c.recursiveNameservers, recursive)
}

func (c *Client) sendQueryCustom(ctx context.Context, fqdn string, rtype uint16, nameservers []string, recursive bool) (*dns.Msg, error) {
	m := createDNSMsg(fqdn, rtype, recursive)

	if len(nameservers) == 0 {
		return nil, &DNSError{Message: "empty list of nameservers"}
	}

	var (
		r      *dns.Msg
		err    error
		errAll error
	)

	for _, ns := range nameservers {
		r, err = c.exchange(ctx, m, ns)
		if err == nil && len(r.Answer) > 0 {
			break
		}

		errAll = errors.Join(errAll, err)
	}

	if err != nil {
		return r, errAll
	}

	return r, nil
}

/*
 * NOTE(ldez): This function is a duplication of `Client.exchange()` from `dns01/client.go`.
 * The 2 functions should be kept in sync.
 */
func (c *Client) exchange(ctx context.Context, m *dns.Msg, ns string) (*dns.Msg, error) {
	if c.tcpOnly {
		r, _, err := c.tcpClient.ExchangeContext(ctx, m, ns)
		if err != nil {
			return r, &DNSError{Message: "DNS call error", MsgIn: m, NS: ns, Err: err}
		}

		return r, nil
	}

	r, _, err := c.udpClient.ExchangeContext(ctx, m, ns)

	if r != nil && r.Truncated {
		// If the TCP request succeeds, the "err" will reset to nil
		r, _, err = c.tcpClient.ExchangeContext(ctx, m, ns)
	}

	if err != nil {
		return r, &DNSError{Message: "DNS call error", MsgIn: m, NS: ns, Err: err}
	}

	return r, nil
}

/*
 * NOTE(ldez): This function is a duplication of `Client.createDNSMsg()` from `dns01/client.go`.
 * The 2 functions should be kept in sync.
 */
func createDNSMsg(fqdn string, rtype uint16, recursive bool) *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, rtype)
	m.SetEdns0(4096, false)

	if !recursive {
		m.RecursionDesired = false
	}

	return m
}
