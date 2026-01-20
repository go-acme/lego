package dns01

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func fakeNS(name, ns string) *dns.NS {
	return &dns.NS{
		Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 172800},
		Ns:  ns,
	}
}

func fakeA(name, ip string) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 10},
		A:   net.ParseIP(ip),
	}
}

func fakeTXT(name, value string) *dns.TXT {
	return &dns.TXT{
		Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 10},
		Txt: []string{value},
	}
}

// mockResolver modifies the default DNS resolver to use a custom network address during the test execution.
// IMPORTANT: it modifying std global variables.
func mockResolver(authoritativeNS net.Addr) func(t *testing.T, client *Client) {
	return func(t *testing.T, client *Client) {
		t.Helper()

		_, port, err := net.SplitHostPort(authoritativeNS.String())
		require.NoError(t, err)

		client.authoritativeNSPort = port

		originalResolver := net.DefaultResolver

		t.Cleanup(func() {
			net.DefaultResolver = originalResolver
		})

		net.DefaultResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 1 * time.Second}

				return d.DialContext(ctx, network, authoritativeNS.String())
			},
		}
	}
}

func mockDefault(t *testing.T, recursiveNS net.Addr, opts ...func(t *testing.T, client *Client)) {
	t.Helper()

	backup := DefaultClient()

	t.Cleanup(func() {
		SetDefaultClient(backup)
	})

	client := NewClient(&Options{RecursiveNameservers: []string{recursiveNS.String()}})

	for _, opt := range opts {
		opt(t, client)
	}

	SetDefaultClient(client)
}
