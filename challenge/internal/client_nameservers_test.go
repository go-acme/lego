package internal

import (
	"sort"
	"testing"

	"github.com/go-acme/lego/v5/challenge"
	dnsmock2 "github.com/go-acme/lego/v5/internal/tester/dnsmock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_LookupAuthoritativeNameservers_OK(t *testing.T) {
	testCases := []struct {
		desc          string
		fakeDNSServer *dnsmock2.Builder
		fqdn          string
		expected      []string
	}{
		{
			fqdn: "en.wikipedia.org.localhost.",
			fakeDNSServer: dnsmock2.NewServer().
				Query("en.wikipedia.org.localhost SOA", dnsmock2.CNAME("dyna.wikimedia.org.localhost")).
				Query("wikipedia.org.localhost SOA", dnsmock2.SOA("")).
				Query("wikipedia.org.localhost NS",
					dnsmock2.Answer(
						fakeNS("wikipedia.org.localhost.", "ns0.wikimedia.org.localhost."),
						fakeNS("wikipedia.org.localhost.", "ns1.wikimedia.org.localhost."),
						fakeNS("wikipedia.org.localhost.", "ns2.wikimedia.org.localhost."),
					),
				),
			expected: []string{"ns0.wikimedia.org.localhost.", "ns1.wikimedia.org.localhost.", "ns2.wikimedia.org.localhost."},
		},
		{
			fqdn: "www.google.com.localhost.",
			fakeDNSServer: dnsmock2.NewServer().
				Query("www.google.com.localhost. SOA", dnsmock2.Noop).
				Query("google.com.localhost. SOA", dnsmock2.SOA("")).
				Query("google.com.localhost. NS",
					dnsmock2.Answer(
						fakeNS("google.com.localhost.", "ns1.google.com.localhost."),
						fakeNS("google.com.localhost.", "ns2.google.com.localhost."),
						fakeNS("google.com.localhost.", "ns3.google.com.localhost."),
						fakeNS("google.com.localhost.", "ns4.google.com.localhost."),
					),
				),
			expected: []string{"ns1.google.com.localhost.", "ns2.google.com.localhost.", "ns3.google.com.localhost.", "ns4.google.com.localhost."},
		},
		{
			fqdn: "mail.proton.me.localhost.",
			fakeDNSServer: dnsmock2.NewServer().
				Query("mail.proton.me.localhost. SOA", dnsmock2.Noop).
				Query("proton.me.localhost. SOA", dnsmock2.SOA("")).
				Query("proton.me.localhost. NS",
					dnsmock2.Answer(
						fakeNS("proton.me.localhost.", "ns1.proton.me.localhost."),
						fakeNS("proton.me.localhost.", "ns2.proton.me.localhost."),
						fakeNS("proton.me.localhost.", "ns3.proton.me.localhost."),
					),
				),
			expected: []string{"ns1.proton.me.localhost.", "ns2.proton.me.localhost.", "ns3.proton.me.localhost."},
		},
	}

	for _, test := range testCases {
		t.Run(test.fqdn, func(t *testing.T) {
			client := NewClient(&Options{RecursiveNameservers: []string{test.fakeDNSServer.Build(t).String()}})

			nss, err := client.LookupAuthoritativeNameservers(t.Context(), test.fqdn)
			require.NoError(t, err)

			sort.Strings(nss)
			sort.Strings(test.expected)

			assert.Equal(t, test.expected, nss)
		})
	}
}

func TestClient_LookupAuthoritativeNameservers_error(t *testing.T) {
	testCases := []struct {
		desc          string
		fqdn          string
		fakeDNSServer *dnsmock2.Builder
		error         string
	}{
		{
			desc: "NXDOMAIN",
			fqdn: "example.invalid.",
			fakeDNSServer: dnsmock2.NewServer().
				Query(". SOA", dnsmock2.Error(dns.RcodeNameError)),
			error: "could not find zone: [fqdn=example.invalid.] could not find the start of authority for 'example.invalid.' [question='invalid. IN  SOA', code=NXDOMAIN]",
		},
		{
			desc: "NS error",
			fqdn: "example.com.",
			fakeDNSServer: dnsmock2.NewServer().
				Query("example.com. SOA", dnsmock2.SOA("")).
				Query("example.com. NS", dnsmock2.Error(dns.RcodeServerFailure)),
			error: "[zone=example.com.] could not determine authoritative nameservers",
		},
		{
			desc: "empty NS",
			fqdn: "example.com.",
			fakeDNSServer: dnsmock2.NewServer().
				Query("example.com. SOA", dnsmock2.SOA("")).
				Query("example.me NS", dnsmock2.Noop),
			error: "[zone=example.com.] could not determine authoritative nameservers",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient(&Options{RecursiveNameservers: []string{test.fakeDNSServer.Build(t).String()}})

			_, err := client.LookupAuthoritativeNameservers(t.Context(), test.fqdn)
			require.Error(t, err)
			assert.EqualError(t, err, test.error)
		})
	}
}

func Test_getNameservers(t *testing.T) {
	testCases := []struct {
		desc     string
		path     string
		stack    challenge.NetworkStack
		expected []string
	}{
		{
			desc:     "with resolv.conf",
			path:     "fixtures/resolv.conf.1",
			stack:    challenge.DualStack,
			expected: []string{"10.200.3.249", "10.200.3.250:5353", "2001:4860:4860::8844", "[10.0.0.1]:5353"},
		},
		{
			desc:     "with nonexistent resolv.conf",
			path:     "fixtures/resolv.conf.nonexistant",
			stack:    challenge.DualStack,
			expected: []string{"1.0.0.1:53", "1.1.1.1:53", "[2606:4700:4700::1001]:53", "[2606:4700:4700::1111]:53"},
		},
		{
			desc:     "default with IPv4Only",
			path:     "resolv.conf.nonexistant",
			stack:    challenge.IPv4Only,
			expected: []string{"1.0.0.1:53", "1.1.1.1:53"},
		},
		{
			desc:     "default with IPv6Only",
			path:     "resolv.conf.nonexistant",
			stack:    challenge.IPv6Only,
			expected: []string{"[2606:4700:4700::1001]:53", "[2606:4700:4700::1111]:53"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			result := getNameservers(test.path, test.stack)

			sort.Strings(result)
			sort.Strings(test.expected)

			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_parseNameservers(t *testing.T) {
	testCases := []struct {
		desc     string
		servers  []string
		expected []string
	}{
		{
			desc:     "without explicit port",
			servers:  []string{"ns1.example.com", "2001:db8::1"},
			expected: []string{"ns1.example.com:53", "[2001:db8::1]:53"},
		},
		{
			desc:     "with explicit port",
			servers:  []string{"ns1.example.com:53", "[2001:db8::1]:53"},
			expected: []string{"ns1.example.com:53", "[2001:db8::1]:53"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := parseNameservers(test.servers)

			assert.Equal(t, test.expected, result)
		})
	}
}
