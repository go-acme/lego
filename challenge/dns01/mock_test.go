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
// IMPORTANT: it modifying global variables.
func mockResolver(t *testing.T, addr net.Addr) {
	t.Helper()

	_, port, err := net.SplitHostPort(addr.String())
	require.NoError(t, err)

	originalDefaultNameserverPort := defaultNameserverPort

	t.Cleanup(func() {
		defaultNameserverPort = originalDefaultNameserverPort
	})

	defaultNameserverPort = port

	originalResolver := net.DefaultResolver

	t.Cleanup(func() {
		net.DefaultResolver = originalResolver
	})

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 1 * time.Second}

			return d.DialContext(ctx, network, addr.String())
		},
	}
}

func useAsNameserver(t *testing.T, addr net.Addr) {
	t.Helper()

	ClearFqdnCache()
	t.Cleanup(func() {
		ClearFqdnCache()
	})

	originalRecursiveNameservers := recursiveNameservers

	t.Cleanup(func() {
		recursiveNameservers = originalRecursiveNameservers
	})

	recursiveNameservers = ParseNameservers([]string{addr.String()})
}
